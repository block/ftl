import type { DeploymentCreatedEvent } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'

export const TimelineDeploymentCreated = ({ deployment }: { deployment: DeploymentCreatedEvent }) => {
  const title = `Deployment created: ${deployment.key}`

  return (
    <span title={title}>
      Created: <span className='text-indigo-500 dark:text-indigo-300'>{deployment.key}</span>
      {deployment.module.length > 0 && (
        <span className='ml-1'>
          with module: <span className='text-emerald-600 dark:text-emerald-400'>{deployment.module}</span>
        </span>
      )}
      {deployment.changeset && (
        <span className='ml-1'>
          with changeset: <span className='text-emerald-600 dark:text-emerald-400'>{deployment.changeset}</span>
        </span>
      )}
    </span>
  )
}
