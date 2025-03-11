import {
  type CallEvent,
  type Event,
  EventType,
  type IngressEvent,
  type PubSubConsumeEvent,
  type PubSubPublishEvent,
} from '../../../protos/xyz/block/ftl/timeline/v1/event_pb.ts'
import type { GetTimelineRequest_Filter } from '../../../protos/xyz/block/ftl/timeline/v1/timeline_pb.ts'
import { eventTypesFilter, requestKeysFilter } from './timeline-filters.ts'
import { useTimeline } from './use-timeline.ts'

export type TraceEvent = CallEvent | IngressEvent | PubSubPublishEvent | PubSubConsumeEvent

export const useRequestTraceEvents = (requestKey?: string, filters: GetTimelineRequest_Filter[] = []) => {
  const eventTypes = [EventType.CALL, EventType.INGRESS, EventType.PUBSUB_CONSUME, EventType.PUBSUB_PUBLISH]

  const allFilters = [...filters, requestKeysFilter([requestKey || '']), eventTypesFilter(eventTypes)]
  const timelineQuery = useTimeline(true, allFilters, 500, !!requestKey)

  const data = (timelineQuery.data?.pages ?? [])
    .flatMap((page): Event[] => (Array.isArray(page) ? page : []))
    .filter(
      (event) =>
        'entry' in event &&
        event.entry &&
        typeof event.entry === 'object' &&
        'case' in event.entry &&
        (event.entry.case === 'call' || event.entry.case === 'ingress' || event.entry.case === 'pubsubPublish' || event.entry.case === 'pubsubConsume'),
    )

  return {
    ...timelineQuery,
    data,
  }
}
