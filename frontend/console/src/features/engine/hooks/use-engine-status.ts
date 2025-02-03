import type { InfiniteData } from '@tanstack/react-query'
import { useMemo } from 'react'
import type { EngineEvent } from '../../../protos/xyz/block/ftl/buildengine/v1/buildengine_pb'
import { compareTimestamps } from '../../../shared/utils'
import { getModuleName } from '../engine.utils'
import { useStreamEngineEvents } from './use-stream-engine-events'

interface EngineStatus {
  engine: EngineEvent | undefined
  modules: Record<string, EngineEvent>
}

const isNewerEvent = (current: EngineEvent, existing: EngineEvent | undefined): boolean => {
  if (!existing) return true
  return compareTimestamps(current.timestamp, existing.timestamp) > 0
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
        // Skip if page is not an array
        if (!Array.isArray(page)) {
          console.warn('Received non-array page in engine status:', page)
          continue
        }

        for (const event of page) {
          if (!event) continue

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
