import { Code, ConnectError } from '@connectrpc/connect'
import { type InfiniteData, useInfiniteQuery, useQueryClient } from '@tanstack/react-query'
import { ConsoleService } from '../../../protos/xyz/block/ftl/console/v1/console_connect'
import type { Event } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { type GetTimelineRequest_Filter, GetTimelineRequest_Order } from '../../../protos/xyz/block/ftl/timeline/v1/timeline_pb'
import { useClient } from '../../../shared/hooks/use-client'
import { useVisibility } from '../../../shared/hooks/use-visibility'

const timelineKey = 'timeline'
const maxTimelineEntries = 1000

export const useTimeline = (isStreaming: boolean, filters: GetTimelineRequest_Filter[], updateIntervalMs = 1000, enabled = true) => {
  const client = useClient(ConsoleService)
  const queryClient = useQueryClient()
  const isVisible = useVisibility()

  const order = GetTimelineRequest_Order.DESC
  const limit = isStreaming ? 200 : 1000

  const queryKey = [timelineKey, isStreaming, filters, order, limit]

  const fetchTimeline = async ({ signal }: { signal: AbortSignal }) => {
    try {
      console.debug('fetching timeline')
      const response = await client.getTimeline({ filters, limit, order }, { signal })
      return response.events
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.Canceled) {
          return []
        }
      }
      throw error
    }
  }

  const streamTimeline = async ({ signal }: { signal: AbortSignal }) => {
    try {
      console.debug('streaming timeline')
      console.debug('timeline-filters:', filters)

      // Initialize with empty pages instead of clearing cache
      queryClient.setQueryData(queryKey, { pages: [], pageParams: [] })

      for await (const response of client.streamTimeline(
        { updateInterval: { seconds: BigInt(0), nanos: updateIntervalMs * 1000 }, query: { limit, filters, order } },
        { signal },
      )) {
        console.debug('timeline-response:', response)
        if (response.events) {
          queryClient.setQueryData<InfiniteData<Event[]>>(queryKey, (old = { pages: [], pageParams: [] }) => {
            const newEvents = response.events
            const existingEvents = old.pages[0] || []
            const uniqueNewEvents = newEvents.filter((newEvent) => !existingEvents.some((existingEvent) => existingEvent.id === newEvent.id))

            // Combine and sort all events by timestamp
            const allEvents = [...uniqueNewEvents, ...existingEvents]
              .sort((a, b) => {
                const aTime = a.timestamp
                const bTime = b.timestamp
                if (!aTime || !bTime) return 0
                return Number(bTime.seconds - aTime.seconds) || Number(bTime.nanos - aTime.nanos)
              })
              .slice(0, maxTimelineEntries)

            return {
              pages: [allEvents, ...old.pages.slice(1)],
              pageParams: old.pageParams,
            }
          })
        }
      }
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code !== Code.Canceled) {
          console.error('Console service - streamEvents - Connect error:', error)
        }
      } else {
        console.error('Console service - streamEvents:', error)
      }
    }
  }

  return useInfiniteQuery({
    queryKey: queryKey,
    queryFn: async ({ signal }) => (isStreaming ? streamTimeline({ signal }) : await fetchTimeline({ signal })),
    enabled: enabled && isVisible,
    getNextPageParam: () => null, // Disable pagination for streaming
    initialPageParam: null, // Disable pagination for streaming
  })
}
