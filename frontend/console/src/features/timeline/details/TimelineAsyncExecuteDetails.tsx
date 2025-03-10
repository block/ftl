import type { AsyncExecuteEvent, Event } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { AttributeBadge } from '../../../shared/components/AttributeBadge'
import { CodeBlockWithTitle } from '../../../shared/components/CodeBlockWithTitle'
import { formatDuration } from '../../../shared/utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../modules/decls/verb/verb.utils'
import { TraceGraph } from '../../traces/TraceGraph'
import { TraceGraphHeader } from '../../traces/TraceGraphHeader'
import { asyncEventTypeString } from '../timeline.utils'

export const TimelineAsyncExecuteDetails = ({ event }: { event: Event }) => {
  const asyncEvent = event.entry.value as AsyncExecuteEvent

  return (
    <>
      <div className='p-4'>
        <div className='pb-2'>
          <TraceGraphHeader requestKey={asyncEvent.requestKey} eventId={event.id} />
          <TraceGraph requestKey={asyncEvent.requestKey} selectedEventId={event.id} />
        </div>

        {asyncEvent.error && <CodeBlockWithTitle title='Error' code={asyncEvent.error} />}
        <DeploymentCard deploymentKey={asyncEvent.deploymentKey} />
        <ul className='pt-4 space-y-2'>
          {asyncEvent.requestKey && (
            <li>
              <AttributeBadge name='request' value={asyncEvent.requestKey} />
            </li>
          )}
          <li>
            <AttributeBadge name='duration' value={formatDuration(asyncEvent.duration)} />
          </li>
          <li>
            <AttributeBadge name='type' value={asyncEventTypeString(asyncEvent.asyncEventType)} />
          </li>
          {asyncEvent.verbRef && (
            <li>
              <AttributeBadge name='destination' value={refString(asyncEvent.verbRef)} />
            </li>
          )}
        </ul>
      </div>
    </>
  )
}
