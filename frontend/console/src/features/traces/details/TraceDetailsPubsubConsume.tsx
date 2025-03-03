import type { Event, PubSubConsumeEvent } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { AttributeBadge } from '../../../shared/components/AttributeBadge'
import { CodeBlockWithTitle } from '../../../shared/components/CodeBlockWithTitle'
import { formatDuration } from '../../../shared/utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'

export const TraceDetailsPubsubConsume = ({ event }: { event: Event }) => {
  const pubsubConsume = event.entry.value as PubSubConsumeEvent
  return (
    <>
      <span className='text-xl font-semibold'>PubSub Publish Details</span>

      {pubsubConsume.error && <CodeBlockWithTitle title='Error' code={pubsubConsume.error} />}

      <DeploymentCard className='mt-4' deploymentKey={pubsubConsume.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        <li>
          <AttributeBadge name='topic' value={pubsubConsume.topic} />
        </li>
        <li>
          <AttributeBadge name='subscription' value={`${pubsubConsume.destVerbModule}.${pubsubConsume.destVerbName}`} />
        </li>
        <li>
          <AttributeBadge name='duration' value={formatDuration(pubsubConsume.duration)} />
        </li>
        {pubsubConsume.requestKey && (
          <li>
            <AttributeBadge name='request' value={pubsubConsume.requestKey} />
          </li>
        )}
      </ul>
    </>
  )
}
