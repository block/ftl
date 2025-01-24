package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"

	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/lsp"
)

type lspCmd struct {
	BuildUpdatesEndpoint *url.URL `help:"Build updates endpoint." default:"http://127.0.0.1:8900" env:"FTL_BUILD_UPDATES_ENDPOINT"`
	languageServer       *lsp.Server
	buildEngineClient    buildenginepbconnect.BuildEngineServiceClient
}

func (l *lspCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx).Scope("lsp")
	logger.Infof("Starting LSP server and listening for updates on %s", l.BuildUpdatesEndpoint)

	g, ctx := errgroup.WithContext(ctx)

	l.languageServer = lsp.NewServer(ctx)
	g.Go(func() error {
		return l.languageServer.Run()
	})

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				logger.Debugf("Connecting to build updates service")
				if err := l.streamBuildEvents(ctx, l.buildEngineClient); err != nil {
					logger.Debugf("Failed to connect to build updates service: %v", err)

					// Delay before reconnecting to avoid tight loop
					time.Sleep(time.Second)
					continue
				}
			}
		}
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error waiting for build events: %w", err)
	}
	return nil
}

func (l *lspCmd) streamBuildEvents(ctx context.Context, client buildenginepbconnect.BuildEngineServiceClient) error {
	stream, err := client.StreamEngineEvents(ctx, connect.NewRequest(&buildenginepb.StreamEngineEventsRequest{}))
	if err != nil {
		return fmt.Errorf("failed to start build events stream: %w", err)
	}

	for stream.Receive() {
		msg := stream.Msg()
		l.languageServer.HandleBuildEvent(ctx, msg)
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("error streaming build events: %w", err)
	}
	return nil
}
