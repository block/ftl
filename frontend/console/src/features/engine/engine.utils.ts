import type { EngineEvent } from '../../protos/xyz/block/ftl/buildengine/v1/buildengine_pb'

export const getModuleName = (event: EngineEvent): string | undefined => {
  switch (event.event.case) {
    case 'moduleAdded':
    case 'moduleRemoved':
      return event.event.value.module

    case 'moduleBuildWaiting':
    case 'moduleBuildStarted':
    case 'moduleBuildFailed':
    case 'moduleBuildSuccess':
      return event.event.value.config?.name

    case 'moduleDeployStarted':
    case 'moduleDeployFailed':
    case 'moduleDeploySuccess':
    case 'moduleDeployWaiting':
      return event.event.value.module

    default:
      return undefined
  }
}

export type ModuleStatus = 'success' | 'error' | 'busy' | 'idle'

export const getEventText = (event: EngineEvent | undefined): string => {
  if (!event?.event) return 'Idle'

  switch (event.event.case) {
    case 'engineStarted':
      return 'Engine started'
    case 'engineEnded': {
      const hasErrors = event.event.value.modules.some((module) => (module.errors?.errors?.length ?? 0) > 0)
      return hasErrors ? 'Engine ended with errors' : 'Engine ended successfully'
    }
    case 'moduleAdded':
      return 'Module added'
    case 'moduleRemoved':
      return 'Module removed'
    case 'moduleBuildWaiting':
      return 'Build waiting'
    case 'moduleBuildStarted':
      return `Build started${event.event.value.isAutoRebuild ? ' (auto rebuild)' : ''}`
    case 'moduleBuildFailed':
      return 'Build failed'
    case 'moduleBuildSuccess':
      return 'Build succeeded'
    case 'moduleDeployStarted':
      return 'Deploy started'
    case 'moduleDeployFailed':
      return 'Deploy failed'
    case 'moduleDeploySuccess':
      return 'Deploy succeeded'
    case 'moduleDeployWaiting':
      return 'Deploy waiting'
    default:
      return `Unknown event ${event.event.case}`
  }
}

export const getModuleStatus = (event: EngineEvent | undefined): ModuleStatus => {
  if (!event) return 'idle'

  // Terminal states should take precedence
  switch (event.event.case) {
    case 'moduleBuildFailed':
    case 'moduleDeployFailed':
      return 'error'
    case 'moduleDeploySuccess':
      return 'success'
    case 'moduleBuildSuccess':
      // Only show success if we're not in the middle of deploying
      return 'success'
    case 'moduleBuildStarted':
    case 'moduleBuildWaiting':
    case 'moduleDeployStarted':
    case 'moduleDeployWaiting':
      return 'busy'
    default:
      return 'idle'
  }
}
