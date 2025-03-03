package rpccontext

import (
	"context"

	"connectrpc.com/connect"
)

type PingResponse[T any] interface {
	*T
	GetNotReady() string
}

type Pingable[Req any, Resp any, RespPtr PingResponse[Resp]] interface {
	Ping(ctx context.Context, req *connect.Request[Req]) (*connect.Response[Resp], error)
}

type clientKey[Client Pingable[Req, Resp, RespPtr], Req any, Resp any, RespPtr PingResponse[Resp]] struct{}

// ContextWithClient returns a context with an RPC client attached.
func ContextWithClient[Client Pingable[Req, Resp, RespPtr], Req any, Resp any, RespPtr PingResponse[Resp]](ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, clientKey[Client, Req, Resp, RespPtr]{}, client)
}

// ClientFromContext returns the given RPC client from the context, or panics.
func ClientFromContext[Client Pingable[Req, Resp, RespPtr], Req any, Resp any, RespPtr PingResponse[Resp]](ctx context.Context) Client {
	value := ctx.Value(clientKey[Client, Req, Resp, RespPtr]{})
	if value == nil {
		panic("no RPC client in context")
	}
	return value.(Client) //nolint:forcetypeassert
}

func IsClientAvailableInContext[Client Pingable[Req, Resp, RespPtr], Req any, Resp any, RespPtr PingResponse[Resp]](ctx context.Context) bool {
	return ctx.Value(clientKey[Client, Req, Resp, RespPtr]{}) != nil
}
