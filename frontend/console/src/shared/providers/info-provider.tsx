import { useQuery } from '@tanstack/react-query'
import type React from 'react'
import { createContext, useContext, useMemo } from 'react'
import { ConsoleService } from '../../protos/xyz/block/ftl/console/v1/console_connect'
import type { GetInfoResponse } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { useClient } from '../hooks/use-client'

interface InfoContextType {
  infoData: GetInfoResponse | null | undefined
  isLoading: boolean
  error: Error | null
  isLocalDev: boolean
}

const getInfoKey = 'getInfo'

const InfoContext = createContext<InfoContextType | undefined>(undefined)

export const InfoProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const client = useClient(ConsoleService)

  const {
    data: infoData,
    isLoading,
    error,
  } = useQuery<GetInfoResponse, Error>({
    queryKey: [getInfoKey],
    queryFn: () => client.getInfo({}),
    staleTime: Number.POSITIVE_INFINITY,
    gcTime: Number.POSITIVE_INFINITY,
  })

  const isLocalDev = useMemo(() => {
    return infoData?.isLocalDev ?? false
  }, [infoData])

  const value: InfoContextType = { infoData, isLoading, error, isLocalDev }

  return <InfoContext.Provider value={value}>{children}</InfoContext.Provider>
}

export const useInfo = (): InfoContextType => {
  const context = useContext(InfoContext)
  if (context === undefined) {
    throw new Error('useInfo must be used within an InfoProvider')
  }
  return context
}
