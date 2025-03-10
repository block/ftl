import type { AsyncExecuteEvent, Event } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { AttributeBadge } from '../../../shared/components/AttributeBadge'
import { CodeBlockWithTitle } from '../../../shared/components/CodeBlockWithTitle'
import { formatDuration } from '../../../shared/utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../modules/decls/verb/verb.utils'
import { asyncEventTypeString } from '../../timeline/timeline.utils'

export const TraceDetailsAsyncCall = ({ event }: { event: Event }) => {
  const asyncCall = event.entry.value as AsyncExecuteEvent

  return (
    <>
      <span className='text-xl font-semibold'>Async Call Details</span>

      {asyncCall.error && <CodeBlockWithTitle title='Error' code={asyncCall.error} />}

      <DeploymentCard className='mt-4' deploymentKey={asyncCall.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        <li>
          <AttributeBadge name='event_type' value={asyncEventTypeString(asyncCall.asyncEventType)} />
        </li>
        <li>
          <AttributeBadge name='duration' value={formatDuration(asyncCall.duration)} />
        </li>
        {asyncCall.requestKey && (
          <li>
            <AttributeBadge name='request' value={asyncCall.requestKey} />
          </li>
        )}
        {asyncCall.verbRef && (
          <li>
            <AttributeBadge name='destination' value={refString(asyncCall.verbRef)} />
          </li>
        )}
        {asyncCall.verbRef && (
          <li>
            <AttributeBadge name='source' value={refString(asyncCall.verbRef)} />
          </li>
        )}
      </ul>
    </>
  )
}
