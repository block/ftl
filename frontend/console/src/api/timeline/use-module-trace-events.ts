import { type Event, EventType } from '../../protos/xyz/block/ftl/timeline/v1/event_pb.ts'
import type { GetTimelineRequest_Filter } from '../../protos/xyz/block/ftl/timeline/v1/timeline_pb.ts'
import { eventTypesFilter, moduleFilter } from './timeline-filters.ts'
import { useTimeline } from './use-timeline.ts'

export const useModuleTraceEvents = (module: string, verb?: string, filters: GetTimelineRequest_Filter[] = []) => {
  const eventTypes = [EventType.CALL, EventType.INGRESS]
  const allFilters = [...filters, moduleFilter(module, verb), eventTypesFilter(eventTypes)]
  const timelineQuery = useTimeline(true, allFilters, 500)

  const data = (timelineQuery.data?.pages ?? [])
    .flatMap((page): Event[] => (Array.isArray(page) ? page : []))
    .filter((event) => 'entry' in event && (event.entry.case === 'call' || event.entry.case === 'ingress'))

  return {
    ...timelineQuery,
    data,
  }
}
