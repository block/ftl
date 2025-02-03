import { Code, ConnectError } from '@connectrpc/connect'
import { useInfiniteQuery, useQueryClient } from '@tanstack/react-query'
import type { InfiniteData } from '@tanstack/react-query'
import type { EngineEvent } from '../../../protos/xyz/block/ftl/buildengine/v1/buildengine_pb'
import { ConsoleService } from '../../../protos/xyz/block/ftl/console/v1/console_connect'
import { useClient } from '../../../shared/hooks/use-client'
import { useVisibility } from '../../../shared/hooks/use-visibility'

const streamEngineEventsKey = 'streamEngineEvents'

export const useStreamEngineEvents = (enabled = true) => {
  const client = useClient(ConsoleService)
  const queryClient = useQueryClient()
  const isVisible = useVisibility()

  const queryKey = [streamEngineEventsKey]

  const streamEngineEvents = async ({ signal }: { signal: AbortSignal }) => {
    try {
      console.debug('streaming engine events')

      queryClient.setQueryData(queryKey, { pages: [[]], pageParams: [0] })

      for await (const response of client.streamEngineEvents({ replayHistory: true }, { signal })) {
        if (response.event) {
          queryClient.setQueryData<InfiniteData<EngineEvent[]>>([streamEngineEventsKey], (old) => {
            const newEvent = response.event as EngineEvent
            if (!old) return { pages: [[newEvent]], pageParams: [0] }

            // Add the new event to the end of the first page to maintain chronological order
            return {
              ...old,
              pages: [[...old.pages[0], newEvent], ...old.pages.slice(1)],
            }
          })
        }
      }
      return queryClient.getQueryData<InfiniteData<EngineEvent[]>>([streamEngineEventsKey]) ?? ({ pages: [[]], pageParams: [0] } as InfiniteData<EngineEvent[]>)
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.Canceled) {
          return (
            queryClient.getQueryData<InfiniteData<EngineEvent[]>>([streamEngineEventsKey]) ?? ({ pages: [[]], pageParams: [0] } as InfiniteData<EngineEvent[]>)
          )
        }
      }
      throw error
    }
  }

  return useInfiniteQuery({
    queryKey: [streamEngineEventsKey],
    queryFn: async ({ signal }) => streamEngineEvents({ signal }),
    enabled: isVisible && enabled,
    getNextPageParam: () => null,
    initialPageParam: 0,
  })
}
