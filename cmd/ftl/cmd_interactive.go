package main

import (
	"context"
	"fmt"

	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/terminal"
)

type interactiveCmd struct {
}

func (i *interactiveCmd) Run(ctx context.Context, binder terminal.KongContextBinder, eventSource *schemaeventsource.EventSource, manager *currentStatusManager) error {
	err := terminal.RunInteractiveConsole(ctx, createKongApplication(&SharedCLI{}, manager), binder, eventSource)
	if err != nil {
		return fmt.Errorf("interactive console: %w", err)
	}
	return nil
}
