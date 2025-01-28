import { Code, ConnectError } from '@connectrpc/connect'
import { useQuery } from '@tanstack/react-query'
import { ConsoleService } from '../../../protos/xyz/block/ftl/console/v1/console_connect'
import { useClient } from '../../../shared/hooks/use-client'

const useStatusKey = 'status'

export const useStatus = () => {
  const client = useClient(ConsoleService)

  const fetchStatus = async (signal: AbortSignal) => {
    try {
      console.debug('fetching status from FTL')
      const status = await client.status({}, { signal })
      return status
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code !== Code.Canceled) {
          console.error('fetchStatus - Connect error:', error)
        }
      } else {
        console.error('fetchStatus:', error)
      }
      throw error
    }
  }

  return useQuery({
    queryKey: [useStatusKey],
    queryFn: async ({ signal }) => fetchStatus(signal),
  })
}
