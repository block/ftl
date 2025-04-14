package rpc

import (
	"context"
	"crypto/tls"
	"iter"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/result"
	"github.com/jpillora/backoff"
	"golang.org/x/net/http2"

	"github.com/block/ftl/internal/authn"
	"github.com/block/ftl/internal/log"
)

// PingResponse is a constraint that is used to enforce that a pointer to the [Pingable] response message has a
// GetNotReady() method.
type PingResponse[T any] interface {
	*T
	GetNotReady() string
}

// Pingable is an interface that is used to indicate that a client can be pinged.
type Pingable[Req any, Resp any, RespPtr PingResponse[Resp]] interface {
	Ping(ctx context.Context, req *connect.Request[Req]) (*connect.Response[Resp], error)
}

// InitialiseClients initialises global HTTP clients used by the RPC system.
//
// "authenticators" are authenticator executables to use for each endpoint. The key is the URL of the endpoint, the
// value is the path to the authenticator executable.
//
// "allowInsecure" skips certificate verification, making TLS susceptible to machine-in-the-middle attacks.
func InitialiseClients(authenticators map[string]string, allowInsecure bool) {
	// We can't have a client-wide timeout because it also applies to
	// streaming RPCs, timing them out.
	h2cClient = &http.Client{
		Transport: authn.Transport(&http2.Transport{
			AllowHTTP: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: allowInsecure, // #nosec G402
			},
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				conn, err := dialer.Dial(network, addr)
				return conn, errors.WithStack(err)
			},
		}, authenticators),
	}
	tlsClient = &http.Client{
		Transport: authn.Transport(&http2.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: allowInsecure, // #nosec G402
			},
			DialTLSContext: func(ctx context.Context, network, addr string, config *tls.Config) (net.Conn, error) {
				tlsDialer := tls.Dialer{Config: config, NetDialer: dialer}
				conn, err := tlsDialer.DialContext(ctx, network, addr)
				return conn, errors.WithStack(err)
			},
		}, authenticators),
	}

	// Use a separate client for HTTP/1.1 with TLS.
	http1TLSClient = &http.Client{
		Transport: authn.Transport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: allowInsecure, // #nosec G402
			},
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				logger := log.FromContext(ctx)
				logger.Debugf("HTTP/1.1 connecting to %s %s", network, addr)

				tlsDialer := tls.Dialer{NetDialer: dialer}
				conn, err := tlsDialer.DialContext(ctx, network, addr)
				return conn, errors.Wrap(err, "HTTP/1.1 TLS dial failed")
			},
		}, authenticators),
	}
}

func init() {
	InitialiseClients(map[string]string{}, false)
}

var (
	dialer = &net.Dialer{
		Timeout: time.Second * 10,
	}
	h2cClient *http.Client
	tlsClient *http.Client
	// Temporary client for HTTP/1.1 with TLS to help with debugging.
	http1TLSClient *http.Client
)

// GetHTTPClient returns a HTTP client usable for the given URL.
func GetHTTPClient(url string) *http.Client {
	if h2cClient == nil {
		panic("rpc.InitialiseClients() must be called before GetHTTPClient()")
	}

	// TEMP_GRPC_HTTP1_ONLY set to non blank will use http1TLSClient
	if os.Getenv("TEMP_GRPC_HTTP1_ONLY") != "" {
		return http1TLSClient
	}

	if strings.HasPrefix(url, "http://") {
		return h2cClient
	}
	return tlsClient
}

// ClientFactory is a function that creates a new client and is typically one of
// the New*Client functions generated by protoc-gen-connect-go.
type ClientFactory[Client Pingable[Req, Resp, RespPtr], Req any, Resp any, RespPtr PingResponse[Resp]] func(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) Client

func Dial[Client Pingable[Req, Resp, RespPtr], Req any, Resp any, RespPtr PingResponse[Resp]](factory ClientFactory[Client, Req, Resp, RespPtr], baseURL string, errorLevel log.Level, opts ...connect.ClientOption) Client {
	client := GetHTTPClient(baseURL)
	opts = append(opts, DefaultClientOptions(errorLevel)...)
	return factory(client, baseURL, opts...)
}

// ContextValuesMiddleware injects values from a Context into the request Context.
func ContextValuesMiddleware(ctx context.Context, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(mergedContext{values: ctx, Context: r.Context()})
		handler.ServeHTTP(w, r)
	}
}

var _ context.Context = (*mergedContext)(nil)

type mergedContext struct {
	values context.Context
	context.Context
}

func (m mergedContext) Value(key any) any {
	if value := m.Context.Value(key); value != nil {
		return value
	}
	return m.values.Value(key)
}

type noopLogSync struct{}

var _ log.Sink = noopLogSync{}

func (noopLogSync) Log(entry log.Entry) error { return nil }

// Wait for a client to become available.
//
// This will repeatedly call Ping() according to the retry policy until the client is
// ready or the deadline is reached.
//
// If "ctx" is cancelled this will return ctx.Err()
//
// Usually rpc errors are logged, but this function will silence ping call errors, and
// returns the last error if the deadline is reached.
func Wait[Req any, Resp any, RespPtr PingResponse[Resp]](ctx context.Context, retry backoff.Backoff, deadline time.Duration, client Pingable[Req, Resp, RespPtr]) error {
	errChan := make(chan error)
	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	go func() {
		logger := log.FromContext(ctx)
		// create a context logger with a new one that does not log debug messages (which include each ping call failures)
		silencedCtx := log.ContextWithLogger(ctx, log.New(log.Error, noopLogSync{}))

		start := time.Now()
		// keep track of the last ping error
		var err error
		for {
			select {
			case <-ctx.Done():
				if err != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
					errChan <- err
				} else {
					errChan <- ctx.Err()
				}
				return
			default:
			}
			var req *Req
			var resp *connect.Response[Resp]
			resp, err = client.Ping(silencedCtx, connect.NewRequest(req))
			if err == nil {
				if (RespPtr)(resp.Msg).GetNotReady() == "" {
					logger.Debugf("Ping succeeded in %.2fs", time.Since(start).Seconds())
					errChan <- nil
					return
				}
				err = errors.Errorf("service is not ready: %s", (RespPtr)(resp.Msg).GetNotReady())
			}
			delay := retry.Duration()
			logger.Tracef("Ping failed waiting %s for client: %+v", delay, err)
			select {
			case <-ctx.Done():
			case <-time.After(delay):
			}
		}
	}()

	err := <-errChan
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// RetryStreamingClientStream will repeatedly call handler with the stream
// returned by "rpc" until handler returns an error or the context is cancelled.
//
// If the stream errors, it will be closed and a new call will be issued.
func RetryStreamingClientStream[Req any, Resp any](
	ctx context.Context,
	retry backoff.Backoff,
	rpc func(context.Context) *connect.ClientStreamForClient[Req, Resp],
	handler func(ctx context.Context, send func(*Req) error) error,
) {
	logLevel := log.Debug
	errored := false
	logger := log.FromContext(ctx)
	for {
		stream := rpc(ctx)
		var err error
		for {
			err = handler(ctx, stream.Send)
			if err != nil {
				break
			}
			if errored {
				logger.Debugf("Client stream recovered")
				errored = false
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
			retry.Reset()
			logLevel = log.Warn
		}

		// We've hit an error.
		_, closeErr := stream.CloseAndReceive()
		if closeErr != nil {
			logger.Logf(log.Debug, "Failed to close stream: %s", closeErr)
		}

		errored = true
		delay := retry.Duration()
		if !errors.Is(err, context.Canceled) {
			logger.Logf(logLevel, "Stream handler failed retrying in %s: %s", delay, err)
		}
		select {
		case <-ctx.Done():
			return

		case <-time.After(delay):
		}

	}
}

// AlwaysRetry instructs RetryStreamingServerStream to always retry the errors it encounters when
// supplied as the errorRetryCallback argument
func AlwaysRetry() func(error) bool {
	return func(err error) bool { return true }
}

// RetryStreamingServerStream will repeatedly call handler with responses from
// the stream returned by "rpc" until either the context is cancelled or the
// errorRetryCallback returns false.
func RetryStreamingServerStream[Req, Resp any](
	ctx context.Context,
	name string,
	retry backoff.Backoff,
	req *Req,
	rpc func(context.Context, *connect.Request[Req]) (*connect.ServerStreamForClient[Resp], error),
	handler func(ctx context.Context, resp *Resp) error,
	errorRetryCallback func(err error) bool,
) {
	logLevel := log.Trace
	errored := false
	logger := log.FromContext(ctx)
	for {
		stream, err := rpc(ctx, connect.NewRequest(req))
		if err == nil {
			for {
				if stream.Receive() {
					resp := stream.Msg()
					err = handler(ctx, resp)

					if err != nil {
						break
					}
					if errored {
						logger.Tracef("Server stream recovered")
						errored = false
					}
					select {
					case <-ctx.Done():
						err := stream.Close()
						if err != nil {
							logger.Debugf("Failed to close stream: %s", err)
						}
						return
					default:
					}
					retry.Reset()
					logLevel = log.Warn
				} else {
					// Stream terminated; check if this was caused by an error
					err = stream.Err()
					logLevel = logLevelForError(err)
					break
				}
			}
			err := stream.Close()
			if err != nil {
				logger.Debugf("Failed to close stream: %s", err)
			}
		}

		errored = true
		delay := retry.Duration()
		if err != nil && !errors.Is(err, context.Canceled) {
			if errorRetryCallback != nil && !errorRetryCallback(err) {
				logger.Errorf(err, "Stream handler encountered a non-retryable error")
				return
			}

			logger.Logf(logLevel, "Stream handler failed for %s, retrying in %s: %s", name, delay, err)
		} else if err == nil {
			logger.Debugf("Stream finished, retrying in %s", delay)
		}

		select {
		case <-ctx.Done():
			return

		case <-time.After(delay):
		}

	}
}

// IterAsGrpc converts an iterator of results into a gRPC stream.
// This is a convenience function to make it easier to unit test streaming endpoints.
//
// If the iterator returns an error, it is returned immediately, and the stream is not sent any more messages.
func IterAsGrpc[T any](iter iter.Seq[result.Result[*T]], stream *connect.ServerStream[T]) error {
	for msg := range iter {
		res, ok := msg.Get()
		if !ok {
			return errors.WithStack(msg.Err()) //nolint:wrapcheck
		}
		if err := stream.Send(res); err != nil {
			return errors.WithStack(err) //nolint:wrapcheck
		}
	}
	return nil
}

// logLevelForError indicates the log.Level to use for the specified error
func logLevelForError(err error) log.Level {
	if err != nil && errors.Is(err, syscall.ECONNREFUSED) {
		return log.Trace
	}
	return log.Warn
}
