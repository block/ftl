package plugin

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/block/ftl/common/log"
	ftlhttp "github.com/block/ftl/internal/http"
	"github.com/block/ftl/internal/local"
	_ "github.com/block/ftl/internal/prodinit"
	"github.com/block/ftl/internal/rpc"
)

type serveCli struct {
	LogConfig log.Config `prefix:"log-" embed:"" group:"Logging:"`
	Bind      *url.URL   `help:"URL to listen on." env:"FTL_BIND" required:""`
	kong.Plugins
}

type serverRegister[Impl any] struct {
	servicePath string
	register    func(i Impl, mux *http.ServeMux)
}

type handlerPath struct {
	path    string
	handler http.Handler
}

type startOptions[Impl any] struct {
	register []serverRegister[Impl]
	handlers []handlerPath
}

// StartOption is an option for Start.
type StartOption[Impl any] func(*startOptions[Impl])

// ConnectHandlerFactory is a type alias for a function that creates a new Connect request handler.
//
// This will typically just be the generated NewXYZHandler function.
type ConnectHandlerFactory[Iface any] func(Iface, ...connect.HandlerOption) (string, http.Handler)

// RegisterAdditionalServer allows a plugin to serve additional gRPC services.
//
// "Impl" must be an implementation of "Iface.
func RegisterAdditionalServer[Impl any, Iface any](servicePath string, register ConnectHandlerFactory[Iface]) StartOption[Impl] {
	return func(so *startOptions[Impl]) {
		so.register = append(so.register, serverRegister[Impl]{
			servicePath: servicePath,
			register: func(i Impl, mux *http.ServeMux) {
				mux.Handle(register(any(i).(Iface), rpc.DefaultHandlerOptions()...)) //nolint:forcetypeassert
			}})
	}
}

// RegisterAdditionalHandler allows a plugin to serve additional HTTP handlers.
func RegisterAdditionalHandler[Impl any](path string, handler http.Handler) StartOption[Impl] {
	return func(so *startOptions[Impl]) {
		so.handlers = append(so.handlers, handlerPath{path: path, handler: handler})
	}
}

// Constructor is a function that creates a new plugin server implementation.
type Constructor[Impl any, Config any] func(context.Context, Config) (context.Context, Impl, error)

// Start a gRPC server plugin listening on the socket specified by the
// environment variable FTL_BIND.
//
// This function does not return.
//
// "Config" is Kong configuration to pass to "create".
// "create" is called to create the implementation of the service.
// "register" is called to register the service with the gRPC server and is typically a generated function.
func Start[Impl any, Iface any, Config any](
	ctx context.Context,
	name string,
	create Constructor[Impl, Config],
	servicePath string,
	register ConnectHandlerFactory[Iface],
	options ...StartOption[Impl],
) {
	var config Config
	cli := serveCli{Plugins: kong.Plugins{&config}}
	kctx := kong.Parse(&cli, kong.Description(`FTL is a platform for building distributed systems that are safe to operate, easy to reason about, and fast to iterate and develop on.`))

	mux := http.NewServeMux()

	so := &startOptions[Impl]{}
	for _, option := range options {
		option(so)
	}

	for _, handler := range so.handlers {
		mux.Handle(handler.path, handler.handler)
	}

	ctx, cancel := context.WithCancelCause(ctx)
	go pollParentExistence(cancel)
	defer cancel(errors.Errorf("plugin %s stopped", name))

	// Configure logging to JSON on stderr. This will be read by the parent process.
	logConfig := cli.LogConfig
	logConfig.JSON = true
	logger := log.Configure(os.Stderr, logConfig)

	logger = logger.Scope(name)
	ctx = log.ContextWithLogger(ctx, logger)

	logger.Tracef("Starting on %s", cli.Bind)

	// Signal handling.
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigch
		logger.Debugf("Terminated by signal %s", sig)
		cancel(errors.Wrapf(context.Canceled, "stopping plugin %s due to signal %s", name, sig))
		// We always kill our children with SIGINT rather than SIGTERM
		// Maven subprocesses will not kill their children correctly if terminated with TERM
		_ = syscall.Kill(-syscall.Getpid(), syscall.SIGINT) //nolint:forcetypeassert,errcheck // best effort
		os.Exit(0)
	}()

	ctx, svc, err := create(ctx, config)
	kctx.FatalIfErrorf(err)

	if _, ok := any(svc).(Iface); !ok {
		var iface Iface
		panic(fmt.Sprintf("%s does not implement %s", reflect.TypeOf(svc), reflect.TypeOf(iface)))
	}

	l, err := net.Listen("tcp", cli.Bind.Host)
	kctx.FatalIfErrorf(err)

	servicePaths := []string{servicePath}

	mux.Handle(register(any(svc).(Iface), rpc.DefaultHandlerOptions()...)) //nolint:forcetypeassert
	for _, register := range so.register {
		register.register(svc, mux)
		servicePaths = append(servicePaths, register.servicePath)
	}

	reflector := grpcreflect.NewStaticReflector(servicePaths...)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	ftlhttp.RegisterPprof(mux)

	// Start the server.
	http1Server := &http.Server{
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: time.Second * 30,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	err = http1Server.Serve(l)
	kctx.FatalIfErrorf(err)

	kctx.Exit(0)
}

func AllocatePort() (*net.TCPAddr, error) {
	addresses, err := local.FreeTCPAddresses(1)
	if err != nil {
		return nil, errors.Wrap(err, "failed to allocate port")
	}
	return addresses[0], nil
}

func cleanup(logger *log.Logger, pidFile string) error {
	pidb, err := os.ReadFile(pidFile)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return errors.WithStack(err)
	}
	if len(pidb) == 0 {
		return nil
	}
	pid, err := strconv.Atoi(string(pidb))
	if err != nil {
		return errors.WithStack(err)
	}
	err = syscall.Kill(pid, syscall.SIGKILL)
	if err != nil && !errors.Is(err, syscall.ESRCH) {
		logger.Warnf("Failed to reap old plugin with pid %d: %s", pid, err)
	}
	return nil
}

func pollParentExistence(cancel context.CancelCauseFunc) {
	// Get the parent process ID
	ppid := os.Getppid()

	for {
		// Check if the parent process is still running
		if !isProcessRunning(ppid) {
			cancel(errors.Errorf("parent process %d is no longer running", ppid))
			break
		}
		// Sleep for a while before checking again
		time.Sleep(1 * time.Second)
	}
}

// isProcessRunning checks if a process with the given PID is running
func isProcessRunning(pid int) bool {
	// Sending signal 0 to a process does not actually send a signal,
	// but it performs error checking and returns an error if the process does not exist.
	err := syscall.Kill(pid, 0)
	return err == nil
}
