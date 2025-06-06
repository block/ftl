package rpc

import (
	"context"
	"net/http"
	"runtime/debug"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"golang.org/x/mod/semver"

	"github.com/block/ftl"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/rpc/headers"
)

type ftlDirectRoutingKey struct{}
type ftlVerbKey struct{}
type requestIDKey struct{}
type parentRequestIDKey struct{}
type transactionIDKey struct{}

// WithDirectRouting ensures any hops in Verb routing do not redirect.
//
// This is used so that eg. calls from Drives do not create recursive loops
// when calling back to the Agent.
func WithDirectRouting(ctx context.Context) context.Context {
	return context.WithValue(ctx, ftlDirectRoutingKey{}, "1")
}

// WithVerbs adds the module.verb chain from the current request to the context.
func WithVerbs(ctx context.Context, verbs []*schema.Ref) context.Context {
	return context.WithValue(ctx, ftlVerbKey{}, verbs)
}

// VerbFromContext returns the current module.verb of the current request.
func VerbFromContext(ctx context.Context) (*schema.Ref, bool) {
	value := ctx.Value(ftlVerbKey{})
	verbs, ok := value.([]*schema.Ref)
	if len(verbs) == 0 {
		return nil, false
	}
	return verbs[len(verbs)-1], ok
}

// VerbsFromContext returns the module.verb chain of the current request.
func VerbsFromContext(ctx context.Context) ([]*schema.Ref, bool) {
	value := ctx.Value(ftlVerbKey{})
	verbs, ok := value.([]*schema.Ref)
	return verbs, ok
}

// IsDirectRouted returns true if the incoming request should be directly
// routed and never redirected.
func IsDirectRouted(ctx context.Context) bool {
	return ctx.Value(ftlDirectRoutingKey{}) != nil
}

// RequestKeyFromContext returns the request key from the context, if any.
//
// TODO: Return an Option here instead of a bool.
func RequestKeyFromContext(ctx context.Context) (optional.Option[key.Request], error) {
	value := ctx.Value(requestIDKey{})
	return errors.WithStack2(requestKeyFromContextValue(value))
}

// WithRequestKey adds the request key to the context.
func WithRequestKey(ctx context.Context, key key.Request) context.Context {
	return context.WithValue(ctx, requestIDKey{}, key.String())
}

func ParentRequestKeyFromContext(ctx context.Context) (optional.Option[key.Request], error) {
	value := ctx.Value(parentRequestIDKey{})
	return errors.WithStack2(requestKeyFromContextValue(value))
}

func WithParentRequestKey(ctx context.Context, key key.Request) context.Context {
	return context.WithValue(ctx, parentRequestIDKey{}, key.String())
}

func requestKeyFromContextValue(value any) (optional.Option[key.Request], error) {
	keyStr, ok := value.(string)
	if !ok {
		return optional.None[key.Request](), nil
	}
	parsedKey, err := key.ParseRequestKey(keyStr)
	if err != nil {
		return optional.None[key.Request](), errors.Wrap(err, "invalid request key")
	}
	return optional.Some(parsedKey), nil
}

func WithTransactionKey(ctx context.Context, key key.TransactionKey) context.Context {
	return context.WithValue(ctx, transactionIDKey{}, key.String())
}

func TransactionKeyFromContext(ctx context.Context) (optional.Option[key.TransactionKey], error) {
	value := ctx.Value(transactionIDKey{})
	return errors.WithStack2(transactionIDFromContextValue(value))
}

func transactionIDFromContextValue(value any) (optional.Option[key.TransactionKey], error) {
	keyStr, ok := value.(string)
	if !ok {
		return optional.None[key.TransactionKey](), nil
	}
	parsedKey, err := key.ParseTransactionKey(keyStr)
	if err != nil {
		return optional.None[key.TransactionKey](), errors.Wrap(err, "invalid transaction key")
	}
	return optional.Some(parsedKey), nil
}

func DefaultClientOptions(level log.Level) []connect.ClientOption {
	interceptors := []connect.Interceptor{
		PanicInterceptor(),
		MetadataInterceptor(log.Trace),
		connectOtelInterceptor(),
		CustomOtelInterceptor(),
	}
	if ftl.Version != "dev" {
		interceptors = append(interceptors, versionInterceptor{})
	}
	return []connect.ClientOption{
		connect.WithGRPC(), // Use gRPC because some servers will not be using Connect.
		connect.WithInterceptors(interceptors...),
	}
}

func DefaultHandlerOptions() []connect.HandlerOption {
	interceptors := []connect.Interceptor{
		PanicInterceptor(),
		MetadataInterceptor(log.Debug),
		connectOtelInterceptor(),
		CustomOtelInterceptor(),
	}
	if ftl.Version != "dev" {
		interceptors = append(interceptors, versionInterceptor{})
	}
	return []connect.HandlerOption{connect.WithInterceptors(interceptors...)}
}

func connectOtelInterceptor() connect.Interceptor {
	otel, err := otelconnect.NewInterceptor(otelconnect.WithTrustRemote(), otelconnect.WithoutServerPeerAttributes())
	if err != nil {
		panic(err)
	}
	return otel
}

// PanicInterceptor intercepts panics and logs them.
func PanicInterceptor() connect.Interceptor {
	return &panicInterceptor{}
}

type panicInterceptor struct{}

// Intercept and log any panics, then re-panic. Defer calls to this function to
// trap panics in the calling function.
func handlePanic(ctx context.Context) {
	logger := log.FromContext(ctx)
	if r := recover(); r != nil {
		var err error
		if rerr, ok := r.(error); ok {
			err = rerr
		} else {
			err = errors.Errorf("%v", r)
		}
		stack := string(debug.Stack())
		logger.Errorf(err, "panic in RPC: %s", stack)
		panic(err)
	}
}

func (*panicInterceptor) WrapStreamingClient(req connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		defer handlePanic(ctx)
		return req(ctx, s)
	}
}

func (*panicInterceptor) WrapStreamingHandler(req connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, s connect.StreamingHandlerConn) error {
		defer handlePanic(ctx)
		return errors.WithStack(req(ctx, s))
	}
}

func (*panicInterceptor) WrapUnary(uf connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		defer handlePanic(ctx)
		return errors.WithStack2(uf(ctx, req))
	}
}

// MetadataInterceptor propagates FTL metadata through servers and clients.
//
// "errorLevel" is the level at which errors will be logged
func MetadataInterceptor(errorLevel log.Level) connect.Interceptor {
	return &metadataInterceptor{
		errorLevel: errorLevel,
	}
}

type metadataInterceptor struct {
	errorLevel log.Level
}

func (*metadataInterceptor) WrapStreamingClient(req connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		// TODO(aat): I can't figure out how to get the client headers here.
		logger := log.FromContext(ctx)
		logger.Tracef("%s (streaming client)", s.Procedure)
		return req(ctx, s)
	}
}

func (m *metadataInterceptor) WrapStreamingHandler(req connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, s connect.StreamingHandlerConn) error {
		logger := log.FromContext(ctx)
		logger.Tracef("%s (streaming handler)", s.Spec().Procedure)
		ctx, err := propagateHeaders(ctx, s.Spec().IsClient, s.RequestHeader())
		if err != nil {
			return errors.WithStack(err)
		}
		err = req(ctx, s)
		if err != nil {
			if connect.CodeOf(err) == connect.CodeCanceled {
				return nil
			}
			logger.Logf(m.errorLevel, "Streaming RPC failed: %s: %s", err, s.Spec().Procedure)
			return errors.WithStack(err)
		}
		return nil
	}
}

func (m *metadataInterceptor) WrapUnary(uf connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		logger := log.FromContext(ctx)
		logger.Tracef("%s (unary)", req.Spec().Procedure)
		ctx, err := propagateHeaders(ctx, req.Spec().IsClient, req.Header())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		resp, err := uf(ctx, req)
		if err != nil {
			logger.Logf(m.errorLevel, "Unary RPC failed: %s: %s", err, req.Spec().Procedure)
			return nil, errors.WithStack(err)
		}
		return resp, nil
	}
}

func propagateHeaders(ctx context.Context, isClient bool, header http.Header) (context.Context, error) {
	if isClient {
		if IsDirectRouted(ctx) {
			headers.SetDirectRouted(header)
		}
		if verbs, ok := VerbsFromContext(ctx); ok {
			headers.SetCallers(header, verbs)
		}
		if key, err := RequestKeyFromContext(ctx); err != nil {
			return nil, errors.WithStack(err)
		} else if key, ok := key.Get(); ok {
			headers.SetRequestKey(header, key)
		}
		if key, err := ParentRequestKeyFromContext(ctx); err != nil {
			return nil, errors.Wrap(err, "invalid parent request key in context")
		} else if key, ok := key.Get(); ok {
			headers.SetParentRequestKey(header, key)
		}
		if txnID, err := TransactionKeyFromContext(ctx); err != nil {
			return nil, errors.WithStack(err)
		} else if txnID, ok := txnID.Get(); ok {
			headers.SetTransactionKey(header, txnID)
		}
	} else {
		if headers.IsDirectRouted(header) {
			ctx = WithDirectRouting(ctx)
		}
		if verbs, err := headers.GetCallers(header); err != nil {
			return nil, errors.WithStack(err)
		} else { //nolint:revive
			ctx = WithVerbs(ctx, verbs)
		}
		if key, ok, err := headers.GetRequestKey(header); err != nil {
			return nil, errors.WithStack(err)
		} else if ok {
			ctx = WithRequestKey(ctx, key)
		}
		if key, ok, err := headers.GetParentRequestKey(header); err != nil {
			return nil, errors.Wrap(err, "invalid parent request key in header")
		} else if ok {
			ctx = WithParentRequestKey(ctx, key)
		}
		if txnID, ok, err := headers.GetTransactionKey(header); err != nil {
			return nil, errors.WithStack(err)
		} else if ok {
			ctx = WithTransactionKey(ctx, txnID)
		}
	}
	return ctx, nil
}

// versionInterceptor reports a warning to the client if the client is older than the server.
type versionInterceptor struct{}

func (v versionInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (v versionInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

func (v versionInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, ar connect.AnyRequest) (connect.AnyResponse, error) {
		resp, err := next(ctx, ar)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if ar.Spec().IsClient {
			if err := v.checkVersion(resp.Header()); err != nil {
				log.FromContext(ctx).Warnf("%s", err)
			}
		} else {
			resp.Header().Set("X-Ftl-Version", ftl.Version)
		}
		return resp, nil
	}
}

func (v versionInterceptor) checkVersion(header http.Header) error {
	version := header.Get("X-Ftl-Version")
	if semver.Compare(ftl.Version, version) < 0 {
		return errors.Errorf("FTL client (%s) is older than server (%s), consider upgrading: https://github.com/block/ftl/releases", ftl.Version, version)
	}
	return nil
}
