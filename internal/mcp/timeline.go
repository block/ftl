package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinepbconnect"
)

type timelineOutput struct {
	Events         []timelineEvent
	NextPageCursor optional.Option[int64]
}

func TimelineTool(ctx context.Context, timelineClient timelinepbconnect.TimelineServiceClient) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool(
			"Timeline",
			mcp.WithDescription(
				`Get the latest runtime events for FTL each:
				- Verb call
				- Ingress call
				- Event published and consumed by pubsub
				- Logs`),
			mcp.WithString("module",
				mcp.Description("Restrict results to a single module"),
				mcp.Pattern(ModuleRegex)),
			mcp.WithString("logLevel",
				mcp.Description("Restrict log results to this log level and above"),
				mcp.Pattern(`(debug|info|warn|error)`),
				mcp.DefaultString("info")),
			mcp.WithString("requestKey",
				mcp.Description("Restrict results to a single request key (can be used to track a single call propagates through the system (including consumption of published events)"),
				mcp.Pattern(`req-[a-z]+-[a-z]+-[a-z0-9]+`)),
			mcp.WithNumber("cursor", mcp.Description("Cursor to paginate results")),
		),
		func(serverCtx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			timelineReq := &timelinepb.GetTimelineRequest{
				Limit: 50,
				Order: timelinepb.GetTimelineRequest_ORDER_DESC,
				Filters: []*timelinepb.GetTimelineRequest_Filter{
					{
						Filter: &timelinepb.GetTimelineRequest_Filter_EventTypes{
							EventTypes: &timelinepb.GetTimelineRequest_EventTypeFilter{
								EventTypes: []timelinepb.EventType{
									timelinepb.EventType_EVENT_TYPE_LOG,
									timelinepb.EventType_EVENT_TYPE_CALL,
									timelinepb.EventType_EVENT_TYPE_INGRESS,
									timelinepb.EventType_EVENT_TYPE_PUBSUB_PUBLISH,
									timelinepb.EventType_EVENT_TYPE_PUBSUB_CONSUME,
								},
							},
						},
					},
				},
			}
			if module, ok := request.Params.Arguments["module"].(string); ok && module != "" {
				timelineReq.Filters = append(timelineReq.Filters, &timelinepb.GetTimelineRequest_Filter{
					Filter: &timelinepb.GetTimelineRequest_Filter_Module{
						Module: &timelinepb.GetTimelineRequest_ModuleFilter{
							Module: module,
						},
					},
				})
			}
			if levelStr, ok := request.Params.Arguments["logLevel"].(string); ok && levelStr != "" {
				var level timelinepb.LogLevel
				switch levelStr {
				case "debug":
					level = timelinepb.LogLevel_LOG_LEVEL_DEBUG
				case "info":
					level = timelinepb.LogLevel_LOG_LEVEL_INFO
				case "warn":
					level = timelinepb.LogLevel_LOG_LEVEL_WARN
				case "error":
					level = timelinepb.LogLevel_LOG_LEVEL_ERROR
				default:
					return nil, fmt.Errorf("invalid log level: %s", levelStr)
				}
				timelineReq.Filters = append(timelineReq.Filters, &timelinepb.GetTimelineRequest_Filter{
					Filter: &timelinepb.GetTimelineRequest_Filter_LogLevel{
						LogLevel: &timelinepb.GetTimelineRequest_LogLevelFilter{
							LogLevel: level,
						},
					},
				})
			}
			if cursor, ok := request.Params.Arguments["cursor"].(int); ok {
				cursor64 := int64(cursor)
				timelineReq.Filters = append(timelineReq.Filters, &timelinepb.GetTimelineRequest_Filter{
					Filter: &timelinepb.GetTimelineRequest_Filter_Id{
						Id: &timelinepb.GetTimelineRequest_IDFilter{
							LowerThan: &cursor64,
						},
					},
				})
			}

			resp, err := timelineClient.GetTimeline(ctx, connect.NewRequest(timelineReq))
			if err != nil {
				return nil, fmt.Errorf("could not get timeline events: %w", err)
			}

			events := make([]timelineEvent, 0, len(resp.Msg.Events))
			for _, event := range resp.Msg.Events {
				out, err := newOutputEvent(event)
				if err != nil {
					return nil, fmt.Errorf("could not convert event: %w", err)
				}
				events = append(events, out)
			}
			output := timelineOutput{
				Events:         events,
				NextPageCursor: optional.Ptr(resp.Msg.Cursor),
			}
			data, err := json.Marshal(output)
			if err != nil {
				return nil, fmt.Errorf("could not marshal results: %w", err)
			}
			return mcp.NewToolResultText(string(data)), nil
		}
}

// timelineEvent is a cleaned up version of a timeline event used for output.
//
// It removes multiple layers of nested messages and converts timestamps to strings.
type timelineEvent map[string]any

func newOutputEvent(raw *timelinepb.Event) (timelineEvent, error) {
	entry := raw.GetEntry()
	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("could not marshal event entry: %w", err)
	}
	var out timelineEvent
	err = json.Unmarshal(entryBytes, &out)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal event entry: %w", err)
	}
	out["timestamp"] = raw.Timestamp.AsTime().Format("2006-01-02 15:04:05:000")
	return out, nil
}
