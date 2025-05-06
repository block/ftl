package app

import (
	"context"
	"io"
	"strings"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/internal/log"
)

type logsCmd struct {
	Follow  bool     `help:"Specify if the logs should be streamed" short:"f"`
	Tail    int      `help:"Number of lines to tail" short:"t" default:"10"`
	Modules []string `arg:"" optional:"" help:"Module names to get logs from. Can be 'module' or 'module:verb' format"`
}

func parseModuleVerb(arg string) (module string, verb string) {
	parts := strings.SplitN(arg, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return arg, ""
}

func (cmd *logsCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	logger := log.FromContext(ctx)

	query := &timelinepb.TimelineQuery{
		Order: timelinepb.TimelineQuery_ORDER_ASC,
		Limit: int32(cmd.Tail), //nolint:gosec
	}

	// Add module filters if modules are specified
	if len(cmd.Modules) > 0 {
		for _, moduleArg := range cmd.Modules {
			module, verb := parseModuleVerb(moduleArg)
			var verbPtr *string
			if verb != "" {
				verbPtr = &verb
			}
			query.Filters = append(query.Filters, &timelinepb.TimelineQuery_Filter{
				Filter: &timelinepb.TimelineQuery_Filter_Module{
					Module: &timelinepb.TimelineQuery_ModuleFilter{
						Module: module,
						Verb:   verbPtr,
					},
				},
			})
		}
	}

	// Stream logs from the server
	stream, err := client.StreamLogs(ctx, connect.NewRequest(&adminpb.StreamLogsRequest{
		Query: query,
	}))
	if err != nil {
		return errors.Wrap(err, "failed to stream logs")
	}

	for stream.Receive() {
		for _, logEvent := range stream.Msg().Logs {
			logger.Log(log.Entry{
				Level:      log.Level(logEvent.LogLevel),
				Message:    logEvent.Message,
				Attributes: logEvent.Attributes,
				Time:       logEvent.Timestamp.AsTime(),
			})
		}

		if !cmd.Follow {
			break
		}
	}

	if err := stream.Err(); err != nil && errors.Is(err, io.EOF) {
		return errors.Wrap(err, "error receiving logs")
	}

	return nil
}
