import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlockWithTitle } from '../../../components/CodeBlockWithTitle'
import type { CallEvent, Event } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../modules/decls/verb/verb.utils'

export const TraceDetailsCall = ({ event }: { event: Event }) => {
  const call = event.entry.value as CallEvent
  return (
    <>
      <span className='text-xl font-semibold'>Call Details</span>
      <CodeBlockWithTitle title='Request' code={JSON.stringify(JSON.parse(call.request), null, 2)} />

      {call.response && <CodeBlockWithTitle title='Response' code={JSON.stringify(JSON.parse(call.response), null, 2)} />}

      {call.error && (
        <>
          <CodeBlockWithTitle title='Error' code={call.error} />
          {call.stack && <CodeBlockWithTitle title='Stack' code={call.stack} />}
        </>
      )}

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
    </>
  )
}
