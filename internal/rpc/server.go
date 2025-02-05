package rpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/alecthomas/concurrency"
	"github.com/alecthomas/types/pubsub"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	gaphttp "github.com/block/ftl/internal/http"
	"github.com/block/ftl/internal/log"
)

const ShutdownGracePeriod = time.Second * 5

type serverOptions struct {
	mux             *http.ServeMux
	reflectionPaths []string
	healthCheck     http.HandlerFunc
	startHooks      []func(ctx context.Context) error
	shutdownHooks   []func(ctx context.Context) error
}

type Option func(*serverOptions)

type GRPCServerConstructor[Iface Pingable[Req, Resp, RespPtr], Req any, Resp any, RespPtr PingResponse[Resp]] func(svc Iface, opts ...connect.HandlerOption) (string, http.Handler)
type RawGRPCServerConstructor[Iface any] func(svc Iface, opts ...connect.HandlerOption) (string, http.Handler)

// GRPC is a convenience function for registering a GRPC server with default options.
// TODO(aat): Do we need pingable here?
func GRPC[Iface, Impl Pingable[Req, Resp, RespPtr], Req any, Resp any, RespPtr PingResponse[Resp]](constructor GRPCServerConstructor[Iface, Req, Resp, RespPtr], impl Impl, options ...connect.HandlerOption) Option {
	return func(o *serverOptions) {
		options = append(options, DefaultHandlerOptions()...)
		path, handler := constructor(any(impl).(Iface), options...)
		o.reflectionPaths = append(o.reflectionPaths, strings.Trim(path, "/"))
		o.mux.Handle(path, handler)
	}
}

// PProf adds /debug/pprof routes to the server.
func PProf() Option {
	return func(so *serverOptions) {
		gaphttp.RegisterPprof(so.mux)
	}
}

func HealthCheck(check http.HandlerFunc) Option {
	return func(options *serverOptions) {
		options.healthCheck = check
	}
}

// RawGRPC is a convenience function for registering a GRPC server with default options without Pingable.
func RawGRPC[Iface, Impl any](constructor RawGRPCServerConstructor[Iface], impl Impl, options ...connect.HandlerOption) Option {
	return func(o *serverOptions) {
		options = append(options, DefaultHandlerOptions()...)
		path, handler := constructor(any(impl).(Iface), options...)
		o.reflectionPaths = append(o.reflectionPaths, strings.Trim(path, "/"))
		o.mux.Handle(path, handler)
	}
}

// HTTP adds a HTTP route to the server.
func HTTP(prefix string, handler http.Handler) Option {
	return func(o *serverOptions) {
		o.mux.Handle(prefix, handler)
	}
}

// ShutdownHook is called when the server is shutting down.
func ShutdownHook(hook func(ctx context.Context) error) Option {
	return func(so *serverOptions) {
		so.shutdownHooks = append(so.shutdownHooks, hook)
	}
}

// StartHook is called when the server is starting up.
func StartHook(hook func(ctx context.Context) error) Option {
	return func(so *serverOptions) {
		so.startHooks = append(so.startHooks, hook)
	}
}

// Options is a convenience function for aggregating multiple options.
func Options(options ...Option) Option {
	return func(so *serverOptions) {
		for _, option := range options {
			option(so)
		}
	}
}

type Server struct {
	listen        *url.URL
	shutdownHooks []func(ctx context.Context) error
	startHooks    []func(ctx context.Context) error
	Bind          *pubsub.Topic[*url.URL] // Will be updated with the actual bind address.
	Server        *http.Server
}

func NewServer(ctx context.Context, listen *url.URL, options ...Option) (*Server, error) {
	opts := &serverOptions{
		mux: http.NewServeMux(),
		healthCheck: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	for _, option := range options {
		option(opts)
	}

	opts.mux.Handle("/healthz", opts.healthCheck)

	// Register reflection services.
	reflector := grpcreflect.NewStaticReflector(opts.reflectionPaths...)
	opts.mux.Handle(grpcreflect.NewHandlerV1(reflector))
	opts.mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	root := ContextValuesMiddleware(ctx, opts.mux)

	http1Server := &http.Server{
		Handler:           h2c.NewHandler(root, &http2.Server{}),
		ReadHeaderTimeout: time.Second * 30,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	return &Server{
		listen:        listen,
		shutdownHooks: opts.shutdownHooks,
		startHooks:    opts.startHooks,
		Bind:          pubsub.New[*url.URL](),
		Server:        http1Server,
	}, nil
}

// Serve runs the server, updating .Bind with the actual bind address.
func (s *Server) Serve(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.listen.Host)
	if err != nil {
		return err
	}
	if s.listen.Port() == "0" {
		s.listen.Host = listener.Addr().String()
	}
	s.Bind.Publish(s.listen)

	tree, _ := concurrency.New(ctx)

	// Shutdown server on context cancellation.
	tree.Go(func(ctx context.Context) error {
		logger := log.FromContext(ctx)

		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), ShutdownGracePeriod)
		defer cancel()
		ctx = log.ContextWithLogger(ctx, logger)

		err := s.Server.Shutdown(ctx)
		if errors.Is(err, context.Canceled) {
			_ = s.Server.Close()
			return err
		}

		for i, hook := range s.shutdownHooks {
			logger.Debugf("Running shutdown hook %d/%d", i+1, len(s.shutdownHooks))
			if err := hook(ctx); err != nil {
				logger.Errorf(err, "shutdown hook failed")
			}
		}

		return nil
	})

	// Start server.
	tree.Go(func(ctx context.Context) error {
		logger := log.FromContext(ctx)
		for i, hook := range s.startHooks {
			logger.Debugf("Running start hook %d/%d", i+1, len(s.startHooks))
			if err := hook(ctx); err != nil {
				logger.Errorf(err, "start hook failed")
			}
		}

		err = s.Server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})

	err = tree.Wait()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Serve starts a HTTP and Connect gRPC server with sane defaults for FTL.
//
// Blocks until the context is cancelled.
func Serve(ctx context.Context, listen *url.URL, options ...Option) error {
	server, err := NewServer(ctx, listen, options...)
	if err != nil {
		return err
	}
	return server.Serve(ctx)
}
