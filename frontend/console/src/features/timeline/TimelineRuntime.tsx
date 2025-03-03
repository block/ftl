import type { DeploymentRuntimeEvent } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'

export const TimelineRuntime = ({ runtime }: { runtime: DeploymentRuntimeEvent }) => {
  const title = `Deployment runtime: ${runtime.key}`

  return (
    <span title={title}>
      {runtime.elementType && <span className='text-purple-500 dark:text-purple-300'>{runtime.elementType}</span>}
      {runtime.elementName && (
        <span className='ml-1'>
          : <span className='text-blue-500 dark:text-blue-300'>{runtime.elementName}</span>
        </span>
      )}{' '}
      {runtime.changeset && (
        <span>
          with changeset: <span className='text-emerald-600 dark:text-emerald-400'>{runtime.changeset}</span>
        </span>
      )}
    </span>
  )
}
