import type { ChangesetStateChangedEvent, Event } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { CodeBlockWithTitle } from '../../../shared/components/CodeBlockWithTitle'

export const TimelineChangesetChangedDetails = ({ event }: { event: Event }) => {
  const changeset = event.entry.value as ChangesetStateChangedEvent

  return (
    <div className='px-4 space-y-4'>
      <div className='mt-4'>
        <CodeBlockWithTitle title='Event Data' code={JSON.stringify(changeset, null, 2)} />
      </div>
    </div>
  )
}
