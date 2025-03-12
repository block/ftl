import type { Timestamp } from '@bufbuild/protobuf'
import type { EventType, LogLevel } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import {
  TimelineQuery_CallFilter,
  TimelineQuery_DeploymentFilter,
  TimelineQuery_EventTypeFilter,
  TimelineQuery_Filter,
  TimelineQuery_IDFilter,
  TimelineQuery_LogLevelFilter,
  TimelineQuery_ModuleFilter,
  TimelineQuery_RequestFilter,
  TimelineQuery_TimeFilter,
} from '../../../protos/xyz/block/ftl/timeline/v1/timeline_pb'

export const requestKeysFilter = (requestKeys: string[]): TimelineQuery_Filter => {
  const filter = new TimelineQuery_Filter()
  const requestFilter = new TimelineQuery_RequestFilter()
  requestFilter.requests = requestKeys
  filter.filter = {
    case: 'requests',
    value: requestFilter,
  }
  return filter
}

export const eventTypesFilter = (eventTypes: EventType[]): TimelineQuery_Filter => {
  const filter = new TimelineQuery_Filter()
  const typesFilter = new TimelineQuery_EventTypeFilter()
  typesFilter.eventTypes = eventTypes
  filter.filter = {
    case: 'eventTypes',
    value: typesFilter,
  }
  return filter
}

export const logLevelFilter = (logLevel: LogLevel): TimelineQuery_Filter => {
  const filter = new TimelineQuery_Filter()
  const logFilter = new TimelineQuery_LogLevelFilter()
  logFilter.logLevel = logLevel
  filter.filter = {
    case: 'logLevel',
    value: logFilter,
  }
  return filter
}

export const modulesFilter = (modules: string[]): TimelineQuery_Filter => {
  const filter = new TimelineQuery_Filter()
  const deploymentsFilter = new TimelineQuery_DeploymentFilter()
  deploymentsFilter.deployments = modules
  filter.filter = {
    case: 'deployments',
    value: deploymentsFilter,
  }
  return filter
}

export const callFilter = (destModule: string, destVerb?: string, sourceModule?: string): TimelineQuery_Filter => {
  const filter = new TimelineQuery_Filter()
  const callFilter = new TimelineQuery_CallFilter()
  callFilter.destModule = destModule
  callFilter.destVerb = destVerb
  callFilter.sourceModule = sourceModule
  filter.filter = {
    case: 'call',
    value: callFilter,
  }
  return filter
}

export const moduleFilter = (module: string, verb?: string): TimelineQuery_Filter => {
  const filter = new TimelineQuery_Filter()
  const moduleFilter = new TimelineQuery_ModuleFilter()
  moduleFilter.module = module
  moduleFilter.verb = verb
  filter.filter = {
    case: 'module',
    value: moduleFilter,
  }
  return filter
}

export const timeFilter = (olderThan: Timestamp | undefined, newerThan: Timestamp | undefined): TimelineQuery_Filter => {
  const filter = new TimelineQuery_Filter()
  const timeFilter = new TimelineQuery_TimeFilter()
  timeFilter.olderThan = olderThan
  timeFilter.newerThan = newerThan
  filter.filter = {
    case: 'time',
    value: timeFilter,
  }
  return filter
}

export const eventIdFilter = ({
  lowerThan,
  higherThan,
}: {
  lowerThan?: bigint
  higherThan?: bigint
}): TimelineQuery_Filter => {
  const filter = new TimelineQuery_Filter()
  const idFilter = new TimelineQuery_IDFilter()
  idFilter.lowerThan = lowerThan
  idFilter.higherThan = higherThan
  filter.filter = {
    case: 'id',
    value: idFilter,
  }
  return filter
}
