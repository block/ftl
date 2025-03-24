package timeline

import (
	"fmt"
	"math"
	"reflect"
	"slices"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	islices "github.com/block/ftl/common/slices"
)

type TimelineFilter func(event *timelinepb.Event) bool

func FilterLogLevel(f *timelinepb.TimelineQuery_LogLevelFilter) TimelineFilter {
	return func(event *timelinepb.Event) bool {
		log, ok := event.Entry.(*timelinepb.Event_Log)
		if !ok {
			// allow non-log events to pass through
			return true
		}
		return int(log.Log.LogLevel) >= int(f.LogLevel)
	}
}

// FilterCall filters call events between the given modules.
//
// Takes a list of filters, with each call event needing to match at least one filter.
func FilterCall(filters []*timelinepb.TimelineQuery_CallFilter) TimelineFilter {
	return func(event *timelinepb.Event) bool {
		call, ok := event.Entry.(*timelinepb.Event_Call)
		if !ok {
			// Allow non-call events to pass through.
			return true
		}
		// Allow event if any event filter matches.
		_, ok = islices.Find(filters, func(f *timelinepb.TimelineQuery_CallFilter) bool {
			if call.Call.DestinationVerbRef.Module != f.DestModule {
				return false
			}
			if f.DestVerb != nil && call.Call.DestinationVerbRef.Name != *f.DestVerb {
				return false
			}
			if f.SourceModule != nil && call.Call.SourceVerbRef.Module != *f.SourceModule {
				return false
			}
			return true
		})
		return ok
	}
}

func FilterModule(filters []*timelinepb.TimelineQuery_ModuleFilter) TimelineFilter {
	return func(event *timelinepb.Event) bool {
		var module, verb string
		switch entry := event.Entry.(type) {
		case *timelinepb.Event_Call:
			module = entry.Call.DestinationVerbRef.Module
			verb = entry.Call.DestinationVerbRef.Name
		case *timelinepb.Event_Ingress:
			module = entry.Ingress.VerbRef.Module
			verb = entry.Ingress.VerbRef.Name
		case *timelinepb.Event_PubsubPublish:
			module = entry.PubsubPublish.VerbRef.Module
			verb = entry.PubsubPublish.VerbRef.Name
		case *timelinepb.Event_PubsubConsume:
			module = *entry.PubsubConsume.DestVerbModule
			verb = *entry.PubsubConsume.DestVerbName
		case *timelinepb.Event_Log:
			module = entry.Log.Attributes["module"]
			verb = entry.Log.Attributes["verb"]
		case *timelinepb.Event_CronScheduled, *timelinepb.Event_ChangesetCreated, *timelinepb.Event_ChangesetStateChanged, *timelinepb.Event_DeploymentRuntime:
			// Block all other event types.
			return false
		default:
			panic(fmt.Sprintf("unexpected event type: %T", event.Entry))
		}
		// Allow event if any module filter matches.
		_, ok := islices.Find(filters, func(f *timelinepb.TimelineQuery_ModuleFilter) bool {
			if f.Module != module {
				return false
			}
			if f.Verb != nil && *f.Verb != verb {
				return false
			}
			return true
		})
		return ok
	}
}

func FilterDeployments(filters []*timelinepb.TimelineQuery_DeploymentFilter) TimelineFilter {
	deployments := islices.Reduce(filters, []string{}, func(acc []string, f *timelinepb.TimelineQuery_DeploymentFilter) []string {
		return append(acc, f.Deployments...)
	})
	return func(event *timelinepb.Event) bool {
		// Always allow changeset events to pass through
		switch event.Entry.(type) {
		case *timelinepb.Event_ChangesetCreated, *timelinepb.Event_ChangesetStateChanged:
			return true
		}

		var deployment string
		switch entry := event.Entry.(type) {
		case *timelinepb.Event_Log:
			deployment = entry.Log.DeploymentKey
		case *timelinepb.Event_Call:
			deployment = entry.Call.DeploymentKey
		case *timelinepb.Event_Ingress:
			deployment = entry.Ingress.DeploymentKey
		case *timelinepb.Event_CronScheduled:
			deployment = entry.CronScheduled.DeploymentKey
		case *timelinepb.Event_PubsubPublish:
			deployment = entry.PubsubPublish.DeploymentKey
		case *timelinepb.Event_PubsubConsume:
			deployment = entry.PubsubConsume.DeploymentKey
		case *timelinepb.Event_DeploymentRuntime:
			deployment = entry.DeploymentRuntime.Key
		default:
			panic(fmt.Sprintf("unexpected event type: %T", event.Entry))
		}
		return slices.Contains(deployments, deployment)
	}
}

func FilterChangesets(filters []*timelinepb.TimelineQuery_ChangesetFilter) TimelineFilter {
	changesets := islices.Reduce(filters, []string{}, func(acc []string, f *timelinepb.TimelineQuery_ChangesetFilter) []string {
		return append(acc, f.Changesets...)
	})
	return func(event *timelinepb.Event) bool {
		switch entry := event.Entry.(type) {
		case *timelinepb.Event_Log:
			if entry.Log.ChangesetKey == nil {
				return false
			}
			return slices.Contains(changesets, *entry.Log.ChangesetKey)
		default:
			return false
		}
	}
}

func FilterRequests(filters []*timelinepb.TimelineQuery_RequestFilter) TimelineFilter {
	requests := islices.Reduce(filters, []string{}, func(acc []string, f *timelinepb.TimelineQuery_RequestFilter) []string {
		return append(acc, f.Requests...)
	})
	return func(event *timelinepb.Event) bool {
		// Always allow changeset events to pass through
		switch event.Entry.(type) {
		case *timelinepb.Event_ChangesetCreated, *timelinepb.Event_ChangesetStateChanged:
			return true
		}

		var request *string
		switch entry := event.Entry.(type) {
		case *timelinepb.Event_Log:
			request = entry.Log.RequestKey
		case *timelinepb.Event_Call:
			request = entry.Call.RequestKey
		case *timelinepb.Event_Ingress:
			request = entry.Ingress.RequestKey
		case *timelinepb.Event_PubsubPublish:
			request = entry.PubsubPublish.RequestKey
		case *timelinepb.Event_PubsubConsume:
			request = entry.PubsubConsume.RequestKey
		case *timelinepb.Event_CronScheduled,
			*timelinepb.Event_DeploymentRuntime,
			*timelinepb.Event_ChangesetCreated,
			*timelinepb.Event_ChangesetStateChanged:
			// These event types don't have request keys
			return false
		}
		if request == nil {
			return false
		}
		return slices.Contains(requests, *request)
	}
}

func FilterTypes(filters ...*timelinepb.TimelineQuery_EventTypeFilter) TimelineFilter {
	types := islices.Reduce(filters, []timelinepb.EventType{}, func(acc []timelinepb.EventType, f *timelinepb.TimelineQuery_EventTypeFilter) []timelinepb.EventType {
		return append(acc, f.EventTypes...)
	})
	allowsAll := slices.Contains(types, timelinepb.EventType_EVENT_TYPE_UNSPECIFIED)
	return func(event *timelinepb.Event) bool {
		if allowsAll {
			return true
		}
		var eventType timelinepb.EventType
		switch event.Entry.(type) {
		case *timelinepb.Event_Log:
			eventType = timelinepb.EventType_EVENT_TYPE_LOG
		case *timelinepb.Event_Call:
			eventType = timelinepb.EventType_EVENT_TYPE_CALL
		case *timelinepb.Event_Ingress:
			eventType = timelinepb.EventType_EVENT_TYPE_INGRESS
		case *timelinepb.Event_CronScheduled:
			eventType = timelinepb.EventType_EVENT_TYPE_CRON_SCHEDULED
		case *timelinepb.Event_PubsubPublish:
			eventType = timelinepb.EventType_EVENT_TYPE_PUBSUB_PUBLISH
		case *timelinepb.Event_PubsubConsume:
			eventType = timelinepb.EventType_EVENT_TYPE_PUBSUB_CONSUME
		case *timelinepb.Event_ChangesetCreated:
			eventType = timelinepb.EventType_EVENT_TYPE_CHANGESET_CREATED
		case *timelinepb.Event_ChangesetStateChanged:
			eventType = timelinepb.EventType_EVENT_TYPE_CHANGESET_STATE_CHANGED
		case *timelinepb.Event_DeploymentRuntime:
			eventType = timelinepb.EventType_EVENT_TYPE_DEPLOYMENT_RUNTIME
		default:
			panic(fmt.Sprintf("unexpected event type: %T", event.Entry))
		}
		return slices.Contains(types, eventType)
	}
}

// FilterTimeRange filters events between the given times, inclusive.
func FilterTimeRange(filter *timelinepb.TimelineQuery_TimeFilter) TimelineFilter {
	return func(event *timelinepb.Event) bool {
		if filter.NewerThan != nil && event.Timestamp.AsTime().Before(filter.NewerThan.AsTime()) {
			return false
		}
		if filter.OlderThan != nil && event.Timestamp.AsTime().After(filter.OlderThan.AsTime()) {
			return false
		}
		return true
	}
}

// FilterIDRange filters events between the given IDs.
func FilterIDRange(filter *timelinepb.TimelineQuery_IDFilter) TimelineFilter {
	return func(event *timelinepb.Event) bool {
		if filter.HigherThan != nil && event.Id <= *filter.HigherThan {
			return false
		}
		if filter.LowerThan != nil && event.Id >= *filter.LowerThan {
			return false
		}
		return true
	}
}

func isAscending(query *timelinepb.TimelineQuery) bool {
	return query.Order != timelinepb.TimelineQuery_ORDER_DESC
}

type filterContext struct {
	lowerThan  int64
	higherThan int64
	filters    []TimelineFilter
}

//nolint:maintidx
func filtersFromQuery(query *timelinepb.TimelineQuery) *filterContext {
	fContext := &filterContext{
		filters:    []TimelineFilter{},
		lowerThan:  math.MaxInt64,
		higherThan: math.MinInt64,
	}

	// Some filters need to be combined (for OR logic), so we group them by type first.
	reqFiltersByType := map[reflect.Type][]*timelinepb.TimelineQuery_Filter{}
	for _, filter := range query.Filters {
		reqFiltersByType[reflect.TypeOf(filter.Filter)] = append(reqFiltersByType[reflect.TypeOf(filter.Filter)], filter)
	}
	if len(reqFiltersByType) == 0 {
		return fContext
	}
	for _, filters := range reqFiltersByType {
		switch filters[0].Filter.(type) {
		case *timelinepb.TimelineQuery_Filter_LogLevel:
			fContext.filters = append(fContext.filters, islices.Map(filters, func(f *timelinepb.TimelineQuery_Filter) TimelineFilter {
				return FilterLogLevel(f.GetLogLevel())
			})...)
		case *timelinepb.TimelineQuery_Filter_Deployments:
			fContext.filters = append(fContext.filters, FilterDeployments(islices.Map(filters, func(f *timelinepb.TimelineQuery_Filter) *timelinepb.TimelineQuery_DeploymentFilter {
				return f.GetDeployments()
			})))
		case *timelinepb.TimelineQuery_Filter_Requests:
			fContext.filters = append(fContext.filters, FilterRequests(islices.Map(filters, func(f *timelinepb.TimelineQuery_Filter) *timelinepb.TimelineQuery_RequestFilter {
				return f.GetRequests()
			})))
		case *timelinepb.TimelineQuery_Filter_EventTypes:
			fContext.filters = append(fContext.filters, FilterTypes(islices.Map(filters, func(f *timelinepb.TimelineQuery_Filter) *timelinepb.TimelineQuery_EventTypeFilter {
				return f.GetEventTypes()
			})...))
		case *timelinepb.TimelineQuery_Filter_Time:
			fContext.filters = append(fContext.filters, islices.Map(filters, func(f *timelinepb.TimelineQuery_Filter) TimelineFilter {
				return FilterTimeRange(f.GetTime())
			})...)
		case *timelinepb.TimelineQuery_Filter_Id:
			fContext.filters = append(fContext.filters, islices.Map(filters, func(f *timelinepb.TimelineQuery_Filter) TimelineFilter {
				return FilterIDRange(f.GetId())
			})...)
			if filters[0].GetId().LowerThan != nil {
				fContext.lowerThan = *filters[0].GetId().LowerThan
			}
			if filters[0].GetId().HigherThan != nil {
				fContext.higherThan = *filters[0].GetId().HigherThan
			}
		case *timelinepb.TimelineQuery_Filter_Call:
			fContext.filters = append(fContext.filters, FilterCall(islices.Map(filters, func(f *timelinepb.TimelineQuery_Filter) *timelinepb.TimelineQuery_CallFilter {
				return f.GetCall()
			})))
		case *timelinepb.TimelineQuery_Filter_Module:
			fContext.filters = append(fContext.filters, FilterModule(islices.Map(filters, func(f *timelinepb.TimelineQuery_Filter) *timelinepb.TimelineQuery_ModuleFilter {
				return f.GetModule()
			})))
		case *timelinepb.TimelineQuery_Filter_Changesets:
			fContext.filters = append(fContext.filters, FilterChangesets(islices.Map(filters, func(f *timelinepb.TimelineQuery_Filter) *timelinepb.TimelineQuery_ChangesetFilter {
				return f.GetChangesets()
			})))
		default:
			panic(fmt.Sprintf("unexpected filter type: %T", filters[0].Filter))
		}
	}
	return fContext
}
