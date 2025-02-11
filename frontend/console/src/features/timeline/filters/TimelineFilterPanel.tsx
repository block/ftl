import { Call02Icon, CustomerServiceIcon, Download04Icon, PackageReceiveIcon, Rocket01Icon, TimeQuarterPassIcon, Upload04Icon } from 'hugeicons-react'
import type React from 'react'
import { useEffect, useState } from 'react'
import { EventType, LogLevel } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import type { GetTimelineRequest_Filter } from '../../../protos/xyz/block/ftl/timeline/v1/timeline_pb'
import { Checkbox } from '../../../shared/components/Checkbox'
import { textColor } from '../../../shared/utils'
import { LogLevelBadgeSmall } from '../../logs/LogLevelBadgeSmall'
import { logLevelBgColor, logLevelColor, logLevelRingColor } from '../../logs/log.utils'
import { useModules } from '../../modules/hooks/use-modules'
import { eventTypesFilter, logLevelFilter, modulesFilter } from '../hooks/timeline-filters'
import { FilterPanelSection } from './FilterPanelSection'

interface EventFilter {
  label: string
  type: EventType
  icon: React.ReactNode
}

const EVENT_TYPES: Record<string, EventFilter> = {
  call: { label: 'Call', type: EventType.CALL, icon: <Call02Icon className='w-4 h-4 text-indigo-500 ml-1' /> },
  asyncExecute: { label: 'Async Call', type: EventType.ASYNC_EXECUTE, icon: <CustomerServiceIcon className='w-4 h-4 text-indigo-500 ml-1' /> },
  log: { label: 'Log', type: EventType.LOG, icon: <LogLevelBadgeSmall logLevel={LogLevel.INFO} /> },
  deploymentCreated: {
    label: 'Deployment Created',
    type: EventType.DEPLOYMENT_CREATED,
    icon: <Rocket01Icon className='w-4 h-4 text-green-500 dark:text-green-300 ml-1' />,
  },
  deploymentUpdated: {
    label: 'Deployment Updated',
    type: EventType.DEPLOYMENT_UPDATED,
    icon: <Rocket01Icon className='w-4 h-4 text-green-500 dark:text-green-300 ml-1' />,
  },
  ingress: { label: 'Ingress', type: EventType.INGRESS, icon: <PackageReceiveIcon className='w-4 h-4 text-sky-400 ml-1' /> },
  cronScheduled: { label: 'Cron Scheduled', type: EventType.CRON_SCHEDULED, icon: <TimeQuarterPassIcon className='w-4 h-4 text-blue-500 ml-1' /> },
  pubsubPublish: { label: 'PubSub Publish', type: EventType.PUBSUB_PUBLISH, icon: <Upload04Icon className='w-4 h-4 text-violet-500 ml-1' /> },
  pubsubConsume: { label: 'PubSub Consume', type: EventType.PUBSUB_CONSUME, icon: <Download04Icon className='w-4 h-4 text-violet-500 ml-1' /> },
}

const LOG_LEVELS: Record<number, string> = {
  1: 'Trace',
  5: 'Debug',
  9: 'Info',
  13: 'Warn',
  17: 'Error',
}

export const TimelineFilterPanel = ({
  onFiltersChanged,
}: {
  onFiltersChanged: (filters: GetTimelineRequest_Filter[]) => void
}) => {
  const modules = useModules()
  const [selectedEventTypes, setSelectedEventTypes] = useState<string[]>(Object.keys(EVENT_TYPES))
  const [selectedModules, setSelectedModules] = useState<string[]>([])
  const [previousModules, setPreviousModules] = useState<string[]>([])
  const [selectedLogLevel, setSelectedLogLevel] = useState<number>(1)

  useEffect(() => {
    if (!modules.isSuccess || modules.data.modules.length === 0) {
      return
    }
    const newModules = modules.data.modules.map((module) => module.runtime?.deployment?.deploymentKey || '')
    const addedModules = newModules.filter((name) => !previousModules.includes(name))

    if (addedModules.length > 0) {
      setSelectedModules((prevSelected) => [...prevSelected, ...addedModules])
    }
    setPreviousModules(newModules)
  }, [modules.data])

  useEffect(() => {
    const filter: GetTimelineRequest_Filter[] = []
    if (selectedEventTypes.length !== Object.keys(EVENT_TYPES).length) {
      const selectedTypes = selectedEventTypes.map((key) => EVENT_TYPES[key].type)

      filter.push(eventTypesFilter(selectedTypes))
    }
    if (selectedLogLevel !== LogLevel.TRACE) {
      filter.push(logLevelFilter(selectedLogLevel))
    }

    filter.push(modulesFilter(selectedModules))

    onFiltersChanged(filter)
  }, [selectedEventTypes, selectedLogLevel, selectedModules])

  const handleTypeChanged = (eventType: string, checked: boolean) => {
    if (checked) {
      setSelectedEventTypes((prev) => [...prev, eventType])
    } else {
      setSelectedEventTypes((prev) => prev.filter((filter) => filter !== eventType))
    }
  }

  const handleModuleChanged = (deploymentKey: string, checked: boolean) => {
    if (checked) {
      setSelectedModules((prev) => [...prev, deploymentKey])
    } else {
      setSelectedModules((prev) => prev.filter((filter) => filter !== deploymentKey))
    }
  }

  const handleLogLevelChanged = (logLevel: string) => {
    setSelectedLogLevel(Number(logLevel))
  }

  return (
    <div className='flex-shrink-0 w-52'>
      <div className='w-full'>
        <div className='mx-auto w-full max-w-md pt-2 pl-2 pb-2'>
          <FilterPanelSection title='Event types'>
            <div className='space-y-1'>
              {Object.keys(EVENT_TYPES).map((key) => (
                <div key={key} className='relative flex items-start group'>
                  <Checkbox
                    id={`event-type-${key}`}
                    checked={selectedEventTypes.includes(key)}
                    onChange={(e) => handleTypeChanged(key, e.target.checked)}
                    label={
                      <div className='flex items-center justify-between w-full relative'>
                        <span className={textColor}>{EVENT_TYPES[key].label}</span>
                        <div className='flex items-center'>
                          <button
                            type='button'
                            onClick={() => setSelectedEventTypes([key])}
                            className='opacity-0 group-hover:opacity-100 text-xs bg-indigo-50 text-indigo-600 ring-1 ring-inset ring-indigo-200 hover:bg-indigo-100 dark:bg-indigo-900 dark:text-indigo-300 dark:hover:bg-indigo-800 dark:ring-1 dark:ring-indigo-800 absolute right-5 rounded px-1 shadow-sm'
                          >
                            only
                          </button>
                          {EVENT_TYPES[key].icon}
                        </div>
                      </div>
                    }
                  />
                </div>
              ))}
              <div className='relative flex items-center pt-1'>
                <button
                  type='button'
                  onClick={() => setSelectedEventTypes(Object.keys(EVENT_TYPES))}
                  className='text-indigo-500 cursor-pointer hover:text-indigo-500'
                >
                  Select all
                </button>
              </div>
            </div>
          </FilterPanelSection>

          <FilterPanelSection title='Log level'>
            <ul className='space-y-1'>
              {Object.keys(LOG_LEVELS).map((key) => (
                <li key={key} onClick={() => handleLogLevelChanged(key)} className='relative flex gap-x-2 cursor-pointer'>
                  <div className='relative flex h-5 w-3 flex-none items-center justify-center'>
                    <div
                      className={`${selectedLogLevel <= Number(key) ? 'h-2.5 w-2.5' : 'h-0.5 w-0.5'} ${
                        selectedLogLevel <= Number(key) ? `${logLevelBgColor[Number(key)]} ${logLevelRingColor[Number(key)]}` : 'bg-gray-300 ring-gray-300'
                      } rounded-full ring-1`}
                    />
                  </div>
                  <p className='flex-auto text-sm leading-5 text-gray-500'>
                    <span className={`${logLevelColor[Number(key)]} flex`}>{LOG_LEVELS[Number(key)]}</span>
                  </p>
                </li>
              ))}
            </ul>
          </FilterPanelSection>

          {modules.isSuccess && (
            <FilterPanelSection title='Modules'>
              {modules.data.modules.map((module) => (
                <div key={module.runtime?.deployment?.deploymentKey} className='group flex items-center'>
                  <Checkbox
                    id={`module-${module.runtime?.deployment?.deploymentKey}`}
                    checked={selectedModules.includes(module.runtime?.deployment?.deploymentKey || '')}
                    onChange={(e) => handleModuleChanged(module.runtime?.deployment?.deploymentKey || '', e.target.checked)}
                    label={<span className={textColor}>{module.name}</span>}
                  />
                  <button
                    type='button'
                    onClick={() => setSelectedModules([module.runtime?.deployment?.deploymentKey || ''])}
                    className='opacity-0 group-hover:opacity-100 text-xs bg-indigo-50 text-indigo-600 ring-1 ring-inset ring-indigo-200 hover:bg-indigo-100 dark:bg-indigo-900 dark:text-indigo-300 dark:hover:bg-indigo-800 dark:ring-1 dark:ring-indigo-800 rounded px-1 shadow-sm ml-auto'
                  >
                    only
                  </button>
                </div>
              ))}
              <div className='relative flex items-center pt-1'>
                <button
                  type='button'
                  onClick={() => setSelectedModules(modules.data.modules.map((module) => module.runtime?.deployment?.deploymentKey || ''))}
                  className='text-indigo-500 cursor-pointer hover:text-indigo-500'
                >
                  Select all
                </button>{' '}
              </div>
            </FilterPanelSection>
          )}
        </div>
      </div>
    </div>
  )
}
