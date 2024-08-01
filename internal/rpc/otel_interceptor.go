package rpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/internal/log"
)

const (
	otelFtlRequestKeyAttr        = attribute.Key("ftl.request_key")
	otelFtlVerbChainAttr         = attribute.Key("ftl.verb_chain")
	otelFtlVerbRefAttr           = attribute.Key("ftl.verb.ref")
	otelFtlVerbModuleAttr        = attribute.Key("ftl.verb.module")
	otelMessageSentSizesAttr     = attribute.Key("ftl.rpc.message.sent.sizes_bytes")
	otelMessageReceivedSizesAttr = attribute.Key("ftl.rpc.message.received.sizes_bytes")
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
		isClient := request.Spec().IsClient

		requestSizesAttr := otelMessageSentSizesAttr
		responseSizesAttr := otelMessageReceivedSizesAttr
		if !isClient {
			requestSizesAttr = otelMessageReceivedSizesAttr
			responseSizesAttr = otelMessageSentSizesAttr
		}

		attributes := getAttributes(ctx)
		requestSize := 0
		if request != nil {
			if msg, ok := request.Any().(proto.Message); ok {
				requestSize = proto.Size(msg)
			}
		}

		response, err := next(ctx, request)
		responseSize := 0
		if err == nil {
			if msg, ok := response.Any().(proto.Message); ok {
				responseSize = proto.Size(msg)
			}
		}

		span := trace.SpanFromContext(ctx)
		span.SetAttributes(append(attributes,
			requestSizesAttr.Int64Slice([]int64{int64(requestSize)}),
			responseSizesAttr.Int64Slice([]int64{int64(responseSize)}),
		)...)
		return response, err
	}
}

func (i *otelInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		attributes := getAttributes(ctx)
		conn := next(ctx, spec)

		state := &streamingState{
			spec:         spec,
			protocol:     conn.Peer().Protocol,
			attributes:   attributes,
			receiveSizes: []int64{},
			sendSizes:    []int64{},
		}

		span := trace.SpanFromContext(ctx)
		return &streamingClientInterceptor{
			StreamingClientConn: conn,
			receive: func(msg any, conn connect.StreamingClientConn) error {
				return state.receive(msg, conn)
			},
			send: func(msg any, conn connect.StreamingClientConn) error {
				return state.send(msg, conn)
			},
			onClose: func() {
				span.SetAttributes(append(state.attributes,
					otelMessageSentSizesAttr.Int64Slice(state.sendSizes),
					otelMessageReceivedSizesAttr.Int64Slice(state.receiveSizes),
				)...)
				if state.error != nil {
					span.SetStatus(codes.Error, state.error.Error())
				}
			},
		}
	}
}

func (i *otelInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		attributes := getAttributes(ctx)
		state := &streamingState{
			spec:         conn.Spec(),
			protocol:     conn.Peer().Protocol,
			attributes:   attributes,
			receiveSizes: []int64{},
			sendSizes:    []int64{},
		}
		streamingHandler := &streamingHandlerInterceptor{
			StreamingHandlerConn: conn,
			receive: func(msg any, conn connect.StreamingHandlerConn) error {
				return state.receive(msg, conn)
			},
			send: func(msg any, conn connect.StreamingHandlerConn) error {
				return state.send(msg, conn)
			},
		}
		err := next(ctx, streamingHandler)
		state.attributes = append(
			state.attributes,
			statusCodeAttribute(conn.Peer().Protocol, err))
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(append(state.attributes,
			otelMessageSentSizesAttr.Int64Slice(state.sendSizes),
			otelMessageReceivedSizesAttr.Int64Slice(state.receiveSizes),
		)...)
		return err
	}
}

func statusCodeAttribute(protocol string, err error) attribute.KeyValue {
	statusCodeKey := fmt.Sprintf("ftl.rpc.%s.status_code", protocol)
	statusCode := attribute.Int64(statusCodeKey, 0)
	if err != nil {
		statusCode = attribute.Int64(statusCodeKey, int64(connect.CodeOf(err)))
	}
	return statusCode
}

// streamingState stores the ongoing metrics for streaming interceptors.
type streamingState struct {
	mu              sync.Mutex
	spec            connect.Spec
	protocol        string
	attributes      []attribute.KeyValue
	error           error
	sentCounter     int64
	receivedCounter int64
	receiveSizes    []int64
	sendSizes       []int64
}

// streamingSenderReceiver encapsulates either a StreamingClientConn or a StreamingHandlerConn.
type streamingSenderReceiver interface {
	Receive(msg any) error
	Send(msg any) error
}

func (s *streamingState) receive(msg any, conn streamingSenderReceiver) error {
	err := conn.Receive(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	if errors.Is(err, io.EOF) {
		return err // nolint:wrapcheck
	}
	if err != nil {
		s.error = err
		s.attributes = append(s.attributes, statusCodeAttribute(s.protocol, err))
	}
	s.receivedCounter++
	if protomsg, ok := msg.(proto.Message); ok {
		size := proto.Size(protomsg)
		s.receiveSizes = append(s.receiveSizes, int64(size))
	}
	return err // nolint:wrapcheck
}

func (s *streamingState) send(msg any, conn streamingSenderReceiver) error {
	err := conn.Send(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	if errors.Is(err, io.EOF) {
		return err // nolint:wrapcheck
	}
	s.sentCounter++
	if err != nil {
		s.error = err
		s.attributes = append(s.attributes, statusCodeAttribute(s.protocol, err))
	}
	if protomsg, ok := msg.(proto.Message); ok {
		size := proto.Size(protomsg)
		s.sendSizes = append(s.sendSizes, int64(size))
	}
	return err // nolint:wrapcheck
}

type streamingClientInterceptor struct {
	connect.StreamingClientConn
	receive func(msg any, conn connect.StreamingClientConn) error
	send    func(any, connect.StreamingClientConn) error
	onClose func()
}

func (s *streamingClientInterceptor) Receive(msg any) error {
	return s.receive(msg, s.StreamingClientConn)
}

func (s *streamingClientInterceptor) Send(msg any) error {
	return s.send(msg, s.StreamingClientConn)
}

func (s *streamingClientInterceptor) Close() error {
	err := s.StreamingClientConn.CloseResponse()
	s.onClose()
	return err // nolint:wrapcheck
}

type streamingHandlerInterceptor struct {
	connect.StreamingHandlerConn
	receive func(any, connect.StreamingHandlerConn) error
	send    func(any, connect.StreamingHandlerConn) error
}

func (s *streamingHandlerInterceptor) Receive(msg any) error {
	return s.receive(msg, s.StreamingHandlerConn)
}

func (s *streamingHandlerInterceptor) Send(msg any) error {
	return s.send(msg, s.StreamingHandlerConn)
}
