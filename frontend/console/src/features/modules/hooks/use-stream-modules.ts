import { Code, ConnectError } from '@connectrpc/connect'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ConsoleService } from '../../../protos/xyz/block/ftl/console/v1/console_connect'
import type { Module, Topology } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { useClient } from '../../../shared/hooks/use-client'
import { useVisibility } from '../../../shared/hooks/use-visibility'

const streamModulesKey = 'streamModules'

export type StreamModulesResult = {
  modules: Module[]
  topology: Topology
}

export const useStreamModules = () => {
  const client = useClient(ConsoleService)
  const queryClient = useQueryClient()
  const isVisible = useVisibility()

  const queryKey = [streamModulesKey]

  const streamModules = async ({ signal }: { signal: AbortSignal }): Promise<StreamModulesResult> => {
    try {
      console.debug('streaming modules')
      let hasModules = false
      for await (const response of client.streamModules({}, { signal })) {
        console.debug('stream-modules-response:', response)
        if (response.modules || response.topology) {
          hasModules = true
          queryClient.setQueryData<StreamModulesResult>(queryKey, (prev = { modules: [], topology: {} as Topology }) => {
            const newModules = response.modules || prev.modules

            return {
              modules: newModules.sort((a, b) => a.name.localeCompare(b.name)),
              topology: response.topology || prev.topology,
            }
          })
        }
      }
      return hasModules ? (queryClient.getQueryData(queryKey) as StreamModulesResult) : { modules: [], topology: {} as Topology }
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code !== Code.Canceled) {
          console.error('Console service - streamModules - Connect error:', error)
        }
      } else {
        console.error('Console service - streamModules:', error)
      }
      return { modules: [], topology: {} as Topology }
    }
  }

  return useQuery<StreamModulesResult>({
    queryKey: queryKey,
    queryFn: ({ signal }) => streamModules({ signal }),
    enabled: isVisible,
  })
}
