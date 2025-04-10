import { DeploymentCard } from '../../../features/deployments/DeploymentCard'
import { TraceGraph } from '../../../features/traces/TraceGraph'
import { TraceGraphHeader } from '../../../features/traces/TraceGraphHeader'
import type { Event, PubSubPublishEvent } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { AttributeBadge } from '../../../shared/components/AttributeBadge'
import { formatDuration } from '../../../shared/utils/date.utils'
import { refString } from '../../modules/decls/verb/verb.utils'

export const TimelinePubSubPublishDetails = ({ event }: { event: Event }) => {
  const pubSubPublish = event.entry.value as PubSubPublishEvent

  return (
    <>
      <div className='p-4'>
        <TraceGraphHeader requestKey={pubSubPublish.requestKey} eventId={event.id} />
        <TraceGraph requestKey={pubSubPublish.requestKey} selectedEventId={event.id} />

        <DeploymentCard className='mt-4' deploymentKey={pubSubPublish.deploymentKey} />

        <ul className='pt-4 space-y-2'>
          <li>
            <AttributeBadge name='origin' value={refString(pubSubPublish.verbRef)} />
          </li>
          <li>
            <AttributeBadge name='topic' value={pubSubPublish.topic} />
          </li>
          <li>
            <AttributeBadge name='duration' value={formatDuration(pubSubPublish.duration)} />
          </li>
        </ul>
      </div>
    </>
  )
}
