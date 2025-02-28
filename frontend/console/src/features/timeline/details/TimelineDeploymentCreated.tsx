import type { DeploymentCreatedEvent, Event } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { CodeBlockWithTitle } from '../../../shared/components/CodeBlockWithTitle'

export const TimelineDeploymentCreatedDetails = ({ event }: { event: Event }) => {
  const deployment = event.entry.value as DeploymentCreatedEvent

  return (
    <div className='px-4 space-y-4'>
      <div className='mt-4'>
        <CodeBlockWithTitle title='Event Data' code={JSON.stringify(deployment, null, 2)} />
      </div>
    </div>
  )
}
