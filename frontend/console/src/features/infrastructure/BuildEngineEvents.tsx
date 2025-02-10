import { useState } from 'react'
import type { EngineEvent } from '../../protos/xyz/block/ftl/buildengine/v1/buildengine_pb'
import type { Error as BuildError } from '../../protos/xyz/block/ftl/language/v1/language_pb'
import { List } from '../../shared/components/List'
import { StatusIndicator } from '../../shared/components/StatusIndicator'
import { formatTimestampShort } from '../../shared/utils'
import { getEventText } from '../engine/engine.utils'

interface BuildEngineEventsProps {
  events: EngineEvent[]
}

const ErrorDisplay = ({ errors }: { errors: BuildError[] }) => {
  const [isExpanded, setIsExpanded] = useState(false)

  if (!errors.length) return null

  return (
    <div className='mt-2 text-xs'>
      <button type='button' onClick={() => setIsExpanded(!isExpanded)} className='text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'>
        {isExpanded ? '▼' : '▶'} {errors.length} error{errors.length !== 1 ? 's' : ''}
      </button>
      {isExpanded && (
        <div className='mt-2 max-h-48 overflow-auto rounded border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800'>
          {errors.map((error, index) => (
            <div key={index} className='p-2 border-b last:border-b-0 border-gray-200 dark:border-gray-700'>
              <div className='text-red-500'>{error.msg}</div>
              {error.pos && (
                <div className='mt-1 text-gray-500'>
                  {error.pos.filename}:{error.pos.line}:{error.pos.startColumn}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

const EventContent = ({ event }: { event: EngineEvent }) => {
  const timestamp = formatTimestampShort(event.timestamp)
  const contextName = (() => {
    switch (event.event.case) {
      case 'moduleAdded':
      case 'moduleRemoved':
        return event.event.value.module
      case 'moduleBuildStarted':
      case 'moduleBuildWaiting':
      case 'moduleBuildSuccess':
      case 'moduleBuildFailed':
        return event.event.value.config?.name ?? 'bogus'
      case 'moduleDeployStarted':
      case 'moduleDeployFailed':
      case 'moduleDeploySuccess':
        return event.event.value.module
      default:
        return 'engine'
    }
  })()

  const getEventStatus = () => {
    switch (event.event.case) {
      case 'moduleBuildSuccess':
      case 'moduleDeploySuccess':
        return 'success'
      case 'moduleBuildFailed':
      case 'moduleDeployFailed':
        return 'error'
      case 'engineEnded':
        const hasErrors = event.event.value.modules.some(module => (module.errors?.errors?.length ?? 0) > 0)
        return hasErrors ? 'error' : 'success'
      default:
        return 'new'
    }
  }

  const getErrors = () => {
    switch (event.event.case) {
      case 'engineEnded':
        return Object.values(event.event.value.modules).flatMap((module) => module.errors?.errors ?? [])
      case 'moduleBuildFailed':
        return event.event.value.errors?.errors ?? []
      default:
        return []
    }
  }

  return (
    <div className='flex items-center w-full'>
      <div className='flex-grow flex items-center gap-x-4'>
        <div className='whitespace-nowrap'>
          <div className='flex items-center gap-x-4'>
            <span className='text-gray-400 text-xs'>{timestamp}</span>
          </div>
          <div className='mt-1'>
            <StatusIndicator state={getEventStatus()} text={getEventText(event)} />
            <ErrorDisplay errors={getErrors()} />
          </div>
        </div>
      </div>

      <div className='flex-none pr-6'>
        <span className='text-xs font-roboto-mono text-gray-500'>{contextName}</span>
      </div>
    </div>
  )
}

export const BuildEngineEvents = ({ events }: BuildEngineEventsProps) => {
  const reversedEvents = [...events].reverse()

  return (
    <div className='overflow-x-hidden w-full'>
      <List items={reversedEvents} renderItem={(event) => <EventContent event={event} />} className='text-xs w-full' />
    </div>
  )
}
