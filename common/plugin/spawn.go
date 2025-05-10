package plugin

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/jpillora/backoff"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

const pluginRetryDelay = time.Millisecond * 50

// PingableClient is a gRPC client that can be pinged.
type PingableClient interface {
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
}

type pluginOptions struct {
	envars            []string
	additionalClients []func(baseURL string, opts ...connect.ClientOption)
	startTimeout      time.Duration
}

// Option used when creating a plugin.
type Option func(*pluginOptions) error

// WithEnvars sets the environment variables to pass to the plugin.
func WithEnvars(envars ...string) Option {
	return func(po *pluginOptions) error {
		po.envars = append(po.envars, envars...)
		return nil
	}
}

// WithStartTimeout sets the timeout for the language-specific drive plugin to start.
func WithStartTimeout(timeout time.Duration) Option {
	return func(po *pluginOptions) error {
		po.startTimeout = timeout
		return nil
	}
}

// WithExtraClient connects to an additional gRPC service in the same plugin.
//
// The client instance is written to "out".
func WithExtraClient[Client rpc.Pingable[Req, Resp, RespPtr], Req any, Resp any, RespPtr rpc.PingResponse[Resp]](out *Client, makeClient rpc.ClientFactory[Client, Req, Resp, RespPtr]) Option {
	return func(po *pluginOptions) error {
		po.additionalClients = append(po.additionalClients, func(baseURL string, opts ...connect.ClientOption) {
			*out = rpc.Dial(makeClient, baseURL, log.Trace, opts...)
		})
		return nil
	}
}

type Plugin[Client rpc.Pingable[Req, Resp, RespPtr], Req any, Resp any, RespPtr rpc.PingResponse[Resp]] struct {
	Cmd      *exec.Cmd
	Endpoint *url.URL // The endpoint the plugin is listening on.
	Client   Client
}

// Spawn a new sub-process plugin.
//
// Plugins are gRPC servers that listen on a socket passed in an envar.
//
// If the subprocess is a Go plugin, it should call [Start] to start the gRPC
// server.
//
// "cmdCtx" will be cancelled when the plugin stops.
//
// The envars passed to the plugin are:
//
//	FTL_BIND - the endpoint URI to listen on
//	FTL_WORKING_DIR - the path to a working directory that the plugin can write state to, if required.
func Spawn[Client rpc.Pingable[Req, Resp, RespPtr], Req any, Resp any, RespPtr rpc.PingResponse[Resp]](
	ctx context.Context,
	defaultLevel log.Level,
	name, module, dir, exe string,
	makeClient rpc.ClientFactory[Client, Req, Resp, RespPtr],
	streamJSON bool,
	options ...Option,
) (plugin *Plugin[Client, Req, Resp, RespPtr], cmdCtx context.Context, err error) {
	logger := log.FromContext(ctx).Scope(name).Module(module)

	opts := pluginOptions{
		startTimeout: time.Second * 30,
	}
	for _, opt := range options {
		if err = opt(&opts); err != nil {
			return nil, nil, errors.WithStack(err)
		}
	}
	workingDir := filepath.Join(dir, ".ftl")
	err = os.Mkdir(workingDir, 0700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, nil, errors.WithStack(err)
	}

	// Clean up previous process.
	pidFile := filepath.Join(workingDir, filepath.Base(exe)+".pid")
	err = cleanup(logger, pidFile)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Find a free port.
	addr, err := AllocatePort()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Start the plugin process.
	pluginEndpoint := &url.URL{Scheme: "http", Host: addr.String()}
	logger.Tracef("Spawning plugin on %s", pluginEndpoint)
	cmd := exec.Command(ctx, defaultLevel, dir, exe)

	if streamJSON {
		// Send the plugin's stderr and stdout to the logger.
		cmd.Stderr = nil
		epipe, err := cmd.StderrPipe()
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create stderr pipe")
		}
		cmd.Stdout = nil
		opipe, err := cmd.StdoutPipe()
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create stdout pipe")
		}

		go func() {
			err := log.JSONStreamer(epipe, logger, log.Error)
			if err != nil {
				logger.Errorf(err, "Error streaming plugin logs.")
			}
		}()
		go func() {
			err := log.JSONStreamer(opipe, logger, log.Info)
			if err != nil {
				logger.Errorf(err, "Error streaming plugin logs.")
			}
		}()
	}
	cmd.Env = append(cmd.Env, "FTL_BIND="+pluginEndpoint.String())
	cmd.Env = append(cmd.Env, "FTL_WORKING_DIR="+workingDir)
	// If the log level is lower than debug we just leave it at the default.
	// Otherwise we set the log level to debug so we can replay it if needed
	// These messages are streamed through this processes logger, so will respect the normal log level
	if logger.GetLevel() >= log.Debug {
		cmd.Env = append(cmd.Env, "LOG_LEVEL=DEBUG")
	} else {
		cmd.Env = append(cmd.Env, "LOG_LEVEL="+logger.GetLevel().String())
	}
	cmd.Env = append(cmd.Env, opts.envars...)
	if err = cmd.Start(); err != nil {
		return nil, nil, errors.WithStack(err)
	}
	// Cancel the context if the command exits - this will terminate the Dial immediately.
	var cancelWithCause context.CancelCauseFunc
	cmdCtx, cancelWithCause = context.WithCancelCause(ctx)
	go func() { cancelWithCause(cmd.Wait()) }()

	defer func() {
		if err != nil {
			logger.Warnf("Plugin failed to start, terminating pid %d", cmd.Process.Pid)
			_ = cmd.Kill(syscall.SIGTERM) //nolint:errcheck // best effort
		}
	}()

	// Write the PID file.
	err = os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0600)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Wait for the plugin to start.
	client := rpc.Dial(makeClient, pluginEndpoint.String(), log.Trace)
	pingErr := make(chan error)
	go func() {
		retry := backoff.Backoff{Min: pluginRetryDelay, Max: pluginRetryDelay}
		err := rpc.Wait[Req, Resp, RespPtr](ctx, retry, opts.startTimeout, client)
		pingErr <- err
		close(pingErr)
	}()

	select {
	case <-cmdCtx.Done():
		return nil, nil, errors.Wrap(cmdCtx.Err(), "plugin process died")

	case err = <-pingErr:
		if err != nil {
			return nil, nil, errors.Wrap(err, "plugin failed to respond to ping")
		}
	}

	for _, makeClient := range opts.additionalClients {
		makeClient(pluginEndpoint.String())
	}

	logger.Debugf("Online")
	plugin = &Plugin[Client, Req, Resp, RespPtr]{Cmd: cmd, Endpoint: pluginEndpoint, Client: client}
	return plugin, cmdCtx, nil
}
