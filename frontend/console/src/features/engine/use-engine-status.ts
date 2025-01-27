import type { InfiniteData } from '@tanstack/react-query'
import { useMemo } from 'react'
import { getModuleName } from '../../features/engine/engine.utils'
import { useStreamEngineEvents } from '../../features/engine/use-stream-engine-events'
import type { EngineEvent } from '../../protos/xyz/block/ftl/buildengine/v1/buildengine_pb'

interface EngineStatus {
  engine: EngineEvent | undefined
  modules: Record<string, EngineEvent>
}

const getEventTimestamp = (event: EngineEvent): bigint => {
  if (!event.timestamp) return BigInt(0)
  return BigInt(event.timestamp.seconds) * BigInt(1e9) + BigInt(event.timestamp.nanos)
}

const isNewerEvent = (current: EngineEvent, existing: EngineEvent | undefined): boolean => {
  if (!existing) return true
  return getEventTimestamp(current) > getEventTimestamp(existing)
}

export const useEngineStatus = (enabled = true) => {
  const events = useStreamEngineEvents(enabled)

  return useMemo<EngineStatus>(() => {
    const status: EngineStatus = {
      engine: undefined,
      modules: {},
    }

    const data = events.data as InfiniteData<EngineEvent[]> | undefined
    if (data?.pages) {
      for (const page of data.pages) {
        for (const event of page) {
          const moduleName = getModuleName(event)
          if (moduleName) {
            // Only update if this is a newer event for this module
            if (isNewerEvent(event, status.modules[moduleName])) {
              status.modules[moduleName] = event
            }
          } else {
            // Only update engine status if this is a newer event
            if (isNewerEvent(event, status.engine)) {
              status.engine = event
            }
          }
        }
      }
    }

    return status
  }, [events.data])
}
