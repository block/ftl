package app

import (
	"context"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/kong"

	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/terminal"
)

type interactiveCmd struct {
}

func (i *interactiveCmd) Run(ctx context.Context, binder KongContextBinder, projectConfig projectconfig.Config, eventSource *schemaeventsource.EventSource, manager *currentStatusManager) error {
	err := terminal.RunInteractiveConsole(ctx, createKongApplication(&InteractiveCLI{}, manager), eventSource, func(ctx context.Context, k *kong.Kong, args []string, additionalExit func(int)) error {
		return errors.WithStack(runInnerCmd(ctx, k, projectConfig, binder, args, additionalExit))
	})
	if err != nil {
		return errors.Wrap(err, "interactive console")
	}
	return nil
}
