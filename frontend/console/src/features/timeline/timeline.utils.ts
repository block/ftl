import { AsyncExecuteEventType } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import type { Event } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'

const eventBackgroundColorMap: Record<string, string> = {
  log: 'bg-gray-500',
  call: 'bg-indigo-500',
  ingress: 'bg-green-500',
  changesetCreated: 'bg-purple-500',
  changesetStateChanged: 'bg-purple-500',
  cronScheduled: 'bg-blue-500',
  asyncExecute: 'bg-indigo-500',
  pubsubPublish: 'bg-teal-500',
  pubsubConsume: 'bg-teal-500',
  deploymentRuntime: 'bg-purple-500',
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
  ingress: 'text-green-500',
  changesetCreated: 'text-purple-500',
  changesetStateChanged: 'text-purple-500',
  cronScheduled: 'text-blue-500',
  asyncExecute: 'text-indigo-500',
  pubsubPublish: 'text-teal-500',
  pubsubConsume: 'text-teal-500',
  deploymentRuntime: 'text-purple-500',
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
