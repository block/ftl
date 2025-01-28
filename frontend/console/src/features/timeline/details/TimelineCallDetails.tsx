import type { CallEvent, Event } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { AttributeBadge } from '../../../shared/components/AttributeBadge'
import { CodeBlockWithTitle } from '../../../shared/components/CodeBlockWithTitle'
import { formatDuration } from '../../../shared/utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../modules/decls/verb/verb.utils'
import { TraceGraph } from '../../traces/TraceGraph'
import { TraceGraphHeader } from '../../traces/TraceGraphHeader'

export const TimelineCallDetails = ({ event }: { event: Event }) => {
  const call = event.entry.value as CallEvent

  return (
    <div className='p-4'>
      <div>
        <TraceGraphHeader requestKey={call.requestKey} eventId={event.id} />
        <TraceGraph requestKey={call.requestKey} selectedEventId={event.id} />
      </div>

      {call.request && <CodeBlockWithTitle title='Request' code={JSON.stringify(JSON.parse(call.request || '{}'), null, 2)} />}

      {call.response && <CodeBlockWithTitle title='Response' code={JSON.stringify(JSON.parse(call.response || '{}'), null, 2)} />}

      {call.error && <CodeBlockWithTitle title='Error' code={call.error} />}

      {call.stack && <CodeBlockWithTitle title='Stack' code={call.stack} />}

      <DeploymentCard className='mt-4' deploymentKey={call.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        {call.requestKey && (
          <li>
            <AttributeBadge name='request' value={call.requestKey} />
          </li>
        )}
        <li>
          <AttributeBadge name='duration' value={formatDuration(call.duration)} />
        </li>
        {call.destinationVerbRef && (
          <li>
            <AttributeBadge name='destination' value={refString(call.destinationVerbRef)} />
          </li>
        )}
        {call.sourceVerbRef && (
          <li>
            <AttributeBadge name='source' value={refString(call.sourceVerbRef)} />
          </li>
        )}
      </ul>
    </div>
  )
}
