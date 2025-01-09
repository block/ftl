package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"

	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	enginepbconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/lsp"
	"github.com/block/ftl/internal/rpc"
)

type lspCmd struct {
	BuildUpdatesEndpoint *url.URL `help:"Build updates endpoint." default:"http://127.0.0.1:8900" env:"FTL_BUILD_UPDATES_ENDPOINT"`
	languageServer       *lsp.Server
}

func (l *lspCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx).Scope("lsp")
	logger.Infof("Starting LSP server and listening for updates on %s", l.BuildUpdatesEndpoint)

	g, ctx := errgroup.WithContext(ctx)

	l.languageServer = lsp.NewServer(ctx)
	g.Go(func() error {
		return l.languageServer.Run()
	})

	client := rpc.Dial(enginepbconnect.NewBuildEngineServiceClient, l.BuildUpdatesEndpoint.String(), log.Error)
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				logger.Debugf("Connecting to build updates service")
				if err := l.streamBuildEvents(ctx, client); err != nil {
					logger.Debugf("Failed to connect to build updates service: %v", err)

					// Delay before reconnecting to avoid tight loop
					time.Sleep(time.Second)
					continue
				}
			}
		}
	})

	return g.Wait()
}

func (l *lspCmd) streamBuildEvents(ctx context.Context, client enginepbconnect.BuildEngineServiceClient) error {
	stream, err := client.StreamEngineEvents(ctx, connect.NewRequest(&buildenginepb.StreamEngineEventsRequest{}))
	if err != nil {
		return fmt.Errorf("failed to start build events stream: %w", err)
	}

	for stream.Receive() {
		msg := stream.Msg()
		l.languageServer.HandleBuildEvent(ctx, msg)
	}

	return stream.Err()
}
