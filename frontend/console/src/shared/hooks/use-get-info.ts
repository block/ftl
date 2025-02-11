import { useQuery } from '@tanstack/react-query'
import { ConsoleService } from '../../protos/xyz/block/ftl/console/v1/console_connect'
import { useClient } from './use-client'

const getInfoKey = 'getInfo'

export const useGetInfo = () => {
  const client = useClient(ConsoleService)

  const { data, isLoading, error } = useQuery({
    queryKey: [getInfoKey],
    queryFn: () => client.getInfo({}),
  })

  return { data, isLoading, error }
}
