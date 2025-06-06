package rpc

import (
	"context"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/block/ftl/common/log"
)

const (
	otelFtlRequestKeyAttr       = attribute.Key("ftl.request_key")
	otelFtlParentRequestKeyAttr = attribute.Key("ftl.parent_request_key")
	otelFtlVerbChainAttr        = attribute.Key("ftl.verb_chain")
	otelFtlVerbRefAttr          = attribute.Key("ftl.verb.ref")
	otelFtlVerbModuleAttr       = attribute.Key("ftl.verb.module")
)

func CustomOtelInterceptor() connect.Interceptor {
	return &otelInterceptor{}
}

type otelInterceptor struct{}

func getAttributes(ctx context.Context) []attribute.KeyValue {
	logger := log.FromContext(ctx)
	attributes := []attribute.KeyValue{}
	requestKey, err := RequestKeyFromContext(ctx)
	if err != nil {
		logger.Warnf("failed to get request key: %s", err)
	}
	if key, ok := requestKey.Get(); ok {
		attributes = append(attributes, otelFtlRequestKeyAttr.String(key.String()))
	}
	parentRequestKey, err := ParentRequestKeyFromContext(ctx)
	if err != nil {
		logger.Warnf("failed to get parent request key: %s", err)
	}
	if key, ok := parentRequestKey.Get(); ok {
		attributes = append(attributes, otelFtlParentRequestKeyAttr.String(key.String()))
	}
	if verb, ok := VerbFromContext(ctx); ok {
		attributes = append(attributes, otelFtlVerbRefAttr.String(verb.String()))
		attributes = append(attributes, otelFtlVerbModuleAttr.String(verb.Module))
	}
	if verbs, ok := VerbsFromContext(ctx); ok {
		verbStrings := make([]string, len(verbs))
		for i, v := range verbs {
			verbStrings[i] = v.String()
		}
		attributes = append(attributes, otelFtlVerbChainAttr.StringSlice(verbStrings))
	}
	return attributes
}

func (i *otelInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		attributes := getAttributes(ctx)
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attributes...)
		return errors.WithStack2(next(ctx, request))
	}
}

func (i *otelInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		attributes := getAttributes(ctx)
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attributes...)
		return next(ctx, spec)
	}
}

func (i *otelInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		attributes := getAttributes(ctx)
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attributes...)
		return errors.WithStack(next(ctx, conn))
	}
}
