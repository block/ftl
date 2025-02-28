import type { ChangesetCreatedEvent } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'

export const TimelineChangesetCreated = ({ changeset }: { changeset: ChangesetCreatedEvent }) => {
  const title = `Changeset created: ${changeset.key} with modules: ${changeset.modules.join(', ')}`

  return (
    <span title={title}>
      Created: <span className='text-indigo-500 dark:text-indigo-300'>{changeset.key}</span>
      {changeset.modules.length > 0 && (
        <span className='ml-1'>
          with modules: <span className='text-emerald-600 dark:text-emerald-400'>{changeset.modules.join(', ')}</span>
        </span>
      )}
      {changeset.toRemove.length > 0 && (
        <span className='ml-1'>
          to remove: <span className='text-red-600 dark:text-red-400'>{changeset.toRemove.join(', ')}</span>
        </span>
      )}
    </span>
  )
}
