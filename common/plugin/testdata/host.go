package main

import (
	"context"
	"time"

	"connectrpc.com/connect"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/internal/log"
)

func main() {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	plugin, _, err := plugin.Spawn(ctx, log.Trace, "test", "test", ".", "./plugin", ftlv1connect.NewVerbServiceClient, plugin.WithStartTimeout(time.Second*2))
	if err != nil {
		panic(err)
	}
	resp, err := plugin.Client.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
	if err != nil {
		panic(err)
	}
	if resp.Msg.NotReady != nil {
		panic(resp.Msg.GetNotReady())
	}
}
