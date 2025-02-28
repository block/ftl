import type { DeploymentRuntimeEvent, Event } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { CodeBlockWithTitle } from '../../../shared/components/CodeBlockWithTitle'

export const TimelineRuntimeDetails = ({ event }: { event: Event }) => {
  const runtime = event.entry.value as DeploymentRuntimeEvent

  return (
    <div className='px-4 space-y-4'>
      <div className='mt-4'>
        <CodeBlockWithTitle title='Event Data' code={JSON.stringify(runtime, null, 2)} />
      </div>
    </div>
  )
}
