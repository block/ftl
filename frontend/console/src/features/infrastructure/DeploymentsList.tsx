import type { Module } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { DeploymentState } from '../../protos/xyz/block/ftl/schema/v1/schema_pb'
import { AttributeBadge } from '../../shared/components/AttributeBadge'
import { Badge } from '../../shared/components/Badge'
import { List } from '../../shared/components/List'
import { classNames } from '../../shared/utils'
import { deploymentTextColor } from '../deployments/deployment.utils'

const getStateBadgeColor = (state: DeploymentState): string => {
  switch (state) {
    case DeploymentState.READY:
    case DeploymentState.CANONICAL:
      return 'bg-green-50 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    case DeploymentState.CANARY:
      return 'bg-yellow-50 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
    case DeploymentState.PROVISIONING:
    case DeploymentState.DE_PROVISIONING:
      return 'bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    case DeploymentState.DRAINING:
      return 'bg-orange-50 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400'
    case DeploymentState.FAILED:
      return 'bg-red-50 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    case DeploymentState.DELETED:
      return 'bg-gray-50 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400'
    default:
      return 'bg-gray-50 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400'
  }
}

const renderAttributeIfPresent = (name: string, value: string | number | undefined | null) => {
  if (value) {
    return <AttributeBadge key={name} name={name} value={value.toString()} />
  }
  return null
}

const renderStateBadge = (state: DeploymentState | undefined | null) => {
  if (!state) return null

  const stateString = DeploymentState[state].toLowerCase().replace(/_/g, ' ').replace('deployment state ', '')

  return <Badge key='state' name={stateString} className={getStateBadgeColor(state)} />
}

export const DeploymentsList = ({ modules }: { modules: Module[] }) => {
  return (
    <List
      items={modules.filter((module) => module.name !== 'builtin')}
      renderItem={(module) => (
        <div className='flex w-full'>
          <div className='flex gap-x-4 items-center w-1/2'>
            <div className='whitespace-nowrap'>
              <div className='flex gap-x-2 items-center'>
                <p>{module.name}</p>
                {renderStateBadge(module.runtime?.deployment?.state)}
              </div>
              {module.runtime?.deployment?.deploymentKey && (
                <p className={classNames(deploymentTextColor(module.runtime?.deployment?.deploymentKey), 'text-sm leading-6')}>
                  {module.runtime?.deployment?.deploymentKey}
                </p>
              )}
            </div>
          </div>
          <div className='w-1/2 flex justify-end'>
            <div className='flex flex-wrap gap-2 justify-end'>
              {renderAttributeIfPresent('min_replicas', module.runtime?.scaling?.minReplicas)}
              {renderAttributeIfPresent('language', module.runtime?.base?.language)}
              {renderAttributeIfPresent('arch', module.runtime?.base?.arch)}
              {renderAttributeIfPresent('os', module.runtime?.base?.os)}
              {renderAttributeIfPresent('created', module.runtime?.base?.createTime?.toDate().toUTCString())}
              {renderAttributeIfPresent('activated', module.runtime?.deployment?.activatedAt?.toDate().toUTCString())}
              {renderAttributeIfPresent('runner', module.runtime?.runner?.endpoint)}
            </div>
          </div>
        </div>
      )}
    />
  )
}
