import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlock } from '../../../components/CodeBlock'
import type { CallEvent, Event } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../modules/decls/verb/verb.utils'

export const TraceDetailsCall = ({ event }: { event: Event }) => {
  const call = event.entry.value as CallEvent
  return (
    <>
      <span className='text-xl font-semibold'>Call Details</span>
      <div className='text-sm pt-2'>Request</div>
      <CodeBlock code={JSON.stringify(JSON.parse(call.request), null, 2)} language='json' />

      {call.response !== 'null' && (
        <>
          <div className='text-sm pt-2'>Response</div>
          <CodeBlock code={JSON.stringify(JSON.parse(call.response), null, 2)} language='json' />
        </>
      )}

      {call.error && (
        <>
          <h3 className='pt-4'>Error</h3>
          <CodeBlock code={call.error} language='text' />
          {call.stack && (
            <>
              <h3 className='pt-4'>Stack</h3>
              <CodeBlock code={call.stack} language='text' />
            </>
          )}
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
