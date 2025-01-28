import type { Event, PubSubPublishEvent } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { AttributeBadge } from '../../../shared/components/AttributeBadge'
import { CodeBlockWithTitle } from '../../../shared/components/CodeBlockWithTitle'
import { formatDuration } from '../../../shared/utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../modules/decls/verb/verb.utils'

export const TraceDetailsPubsubPublish = ({ event }: { event: Event }) => {
  const pubsubPublish = event.entry.value as PubSubPublishEvent
  return (
    <>
      <span className='text-xl font-semibold'>Publish Details</span>

      {pubsubPublish.request && <CodeBlockWithTitle title='Request' code={pubsubPublish.request} />}
      {pubsubPublish.error && <CodeBlockWithTitle title='Error' code={pubsubPublish.error} />}

      <DeploymentCard className='mt-4' deploymentKey={pubsubPublish.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        <li>
          <AttributeBadge name='topic' value={pubsubPublish.topic} />
        </li>
        <li>
          <AttributeBadge name='duration' value={formatDuration(pubsubPublish.duration)} />
        </li>
        {pubsubPublish.requestKey && (
          <li>
            <AttributeBadge name='request' value={pubsubPublish.requestKey} />
          </li>
        )}
        {pubsubPublish.verbRef && (
          <li>
            <AttributeBadge name='verb_ref' value={refString(pubsubPublish.verbRef)} />
          </li>
        )}
      </ul>
    </>
  )
}
