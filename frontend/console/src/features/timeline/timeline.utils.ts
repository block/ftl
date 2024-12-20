import { AsyncExecuteEventType } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import type { Event } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'

const eventBackgroundColorMap: Record<string, string> = {
  log: 'bg-gray-500',
  call: 'bg-indigo-500',
  ingress: 'bg-sky-400',
  deploymentCreated: 'bg-green-500 dark:bg-green-300',
  deploymentUpdated: 'bg-green-500 dark:bg-green-300',
  cronScheduled: 'bg-blue-500',
  asyncExecute: 'bg-indigo-500',
  pubsubPublish: 'bg-violet-500',
  pubsubConsume: 'bg-violet-500',
  '': 'bg-gray-500',
}

export const eventBackgroundColor = (event: Event) => {
  if (isError(event)) {
    return 'bg-red-500'
  }
  return eventBackgroundColorMap[event.entry.case || '']
}

const eventTextColorMap: Record<string, string> = {
  log: 'text-gray-500',
  call: 'text-indigo-500',
  ingress: 'text-sky-400',
  deploymentCreated: 'text-green-500 dark:text-green-300',
  deploymentUpdated: 'text-green-500 dark:text-green-300',
  cronScheduled: 'text-blue-500',
  asyncExecute: 'text-indigo-500',
  pubsubPublish: 'text-violet-500',
  pubsubConsume: 'text-violet-500',
  '': 'text-gray-500',
}

export const asyncEventTypeString = (type: AsyncExecuteEventType) => {
  switch (type) {
    case AsyncExecuteEventType.CRON:
      return 'cron'
    case AsyncExecuteEventType.PUBSUB:
      return 'pubsub'
    default:
      return 'unknown'
  }
}

export const eventTextColor = (event: Event) => {
  if (isError(event)) {
    return 'text-red-500'
  }
  return eventTextColorMap[event.entry.case || '']
}

const isError = (event: Event) => {
  if (event.entry.case === 'call' && event.entry.value.error) {
    return true
  }
  if (event.entry.case === 'ingress' && event.entry.value.error) {
    return true
  }
  if (event.entry.case === 'asyncExecute' && event.entry.value.error) {
    return true
  }
  return false
}
