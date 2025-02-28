import { ChangesetState } from '../../protos/xyz/block/ftl/schema/v1/schema_pb'
import type { ChangesetStateChangedEvent } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'

export const TimelineChangesetChanged = ({ changeset }: { changeset: ChangesetStateChangedEvent }) => {
  const title = `Changeset state changed: ${changeset.key} to ${ChangesetState[changeset.state]}${changeset.error ? ` with error: ${changeset.error}` : ''}`

  // Determine state color based on the state value
  const getStateColor = (state: ChangesetState) => {
    switch (state) {
      case ChangesetState.FINALIZED:
        return 'text-green-600 dark:text-green-400'
      case ChangesetState.FAILED:
        return 'text-red-600 dark:text-red-400'
      case ChangesetState.PREPARING:
      case ChangesetState.PREPARED:
      case ChangesetState.COMMITTED:
      case ChangesetState.DRAINED:
        return 'text-blue-600 dark:text-blue-400'
      case ChangesetState.ROLLING_BACK:
        return 'text-yellow-600 dark:text-yellow-400'
      default:
        return 'text-gray-600 dark:text-gray-400'
    }
  }

  const stateColor = getStateColor(changeset.state)
  // Remove the CHANGESET_STATE_ prefix for display
  const displayState = ChangesetState[changeset.state].replace('CHANGESET_STATE_', '').toLowerCase()

  return (
    <span title={title}>
      <span className={stateColor}>{displayState}</span>
      {changeset.error && (
        <span className='ml-1'>
          with error: <span className='text-red-600 dark:text-red-400'>{changeset.error}</span>
        </span>
      )}
    </span>
  )
}
