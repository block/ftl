package main

import (
	"context"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/jpillora/backoff"

	timelinev1 "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/terminal"
	"github.com/block/ftl/internal/timelineclient"
)

type replayCmd struct {
	Wait            time.Duration  `short:"w" help:"Wait up to this elapsed time for the FTL cluster to become available." default:"1m"`
	Verb            reflection.Ref `arg:"" required:"" help:"Full path of Verb to call." predictor:"verbs"`
	Verbose         bool           `flag:"" short:"v" help:"Print verbose information."`
	ConsoleEndpoint *url.URL       `help:"Console endpoint." env:"FTL_CONTROLLER_CONSOLE_URL" default:"http://127.0.0.1:8892"`
}

func (c *replayCmd) Run(
	ctx context.Context,
	verbClient ftlv1connect.VerbServiceClient,
	eventSource *schemaeventsource.EventSource,
	timelineClient *timelineclient.Client,
) error {
	// Wait timeout is for both pings to complete, not each ping individually
	startTime := time.Now()

	if err := rpc.Wait(ctx, backoff.Backoff{Max: time.Second * 2}, c.Wait, verbClient); err != nil {
		return errors.Wrap(err, "failed to wait for client")
	}

	if err := rpc.Wait(ctx, backoff.Backoff{Max: time.Second * 2}, c.Wait-time.Since(startTime), timelineClient); err != nil {
		return errors.Wrap(err, "failed to wait for console service client")
	}

	logger := log.FromContext(ctx)

	// First check the verb is valid
	// lookup the verbs
	eventSource.WaitForInitialSync(ctx)

	found := false
	for _, module := range eventSource.CanonicalView().InternalModules() {
		if module.Name == c.Verb.Module {
			for _, v := range module.Verbs() {
				if v.Name == c.Verb.Name {
					found = true
					break
				}
			}
		}
	}
	if !found {
		suggestions, err := findSuggestions(ctx, eventSource, c.Verb)
		// if we have suggestions, return a helpful error message. otherwise continue to the original error
		if err == nil {
			return errors.Errorf("verb not found: %s\n\nDid you mean one of these?\n%s", c.Verb, strings.Join(suggestions, "\n"))
		}
		return errors.Errorf("verb not found: %s", c.Verb)
	}

	events, err := timelineClient.GetTimeline(ctx, connect.NewRequest(&timelinev1.GetTimelineRequest{
		Query: &timelinev1.TimelineQuery{
			Order: timelinev1.TimelineQuery_ORDER_DESC,
			Filters: []*timelinev1.TimelineQuery_Filter{
				{
					Filter: &timelinev1.TimelineQuery_Filter_Call{
						Call: &timelinev1.TimelineQuery_CallFilter{
							DestModule: c.Verb.Module,
							DestVerb:   &c.Verb.Name,
						},
					},
				},
				{
					Filter: &timelinev1.TimelineQuery_Filter_EventTypes{
						EventTypes: &timelinev1.TimelineQuery_EventTypeFilter{
							EventTypes: []timelinev1.EventType{timelinev1.EventType_EVENT_TYPE_CALL},
						},
					},
				},
			},
		},
	}))
	if err != nil {
		return errors.Wrap(err, "failed to get events")
	}
	if len(events.Msg.GetEvents()) == 0 {
		return errors.Errorf("no events found for %v", c.Verb)
	}
	requestJSON := events.Msg.GetEvents()[0].GetCall().Request

	logger.Debugf("Replaying %s with body:", c.Verb)
	terminal.PrintJSON(ctx, []byte(requestJSON))
	logger.Debugf("Response:")

	// Create a callCmd with the same console endpoint configuration
	cmd := &callCmd{
		ConsoleEndpoint: c.ConsoleEndpoint,
		Verbose:         c.Verbose,
	}
	return errors.WithStack(callVerb(ctx, verbClient, eventSource, c.Verb, []byte(requestJSON), c.Verbose, cmd))
}
