import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlockWithTitle } from '../../../components/CodeBlockWithTitle'
import type { Event, LogEvent } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { DeploymentCard } from '../../deployments/DeploymentCard'

export const TimelineLogDetails = ({ log }: { event: Event; log: LogEvent }) => {
  return (
    <div className='px-4'>
      {log.message && <CodeBlockWithTitle title='Message' code={log.message} />}

      {log.attributes && <CodeBlockWithTitle title='Attributes' code={JSON.stringify(log.attributes, null, 2)} />}

      <DeploymentCard className='mt-4' deploymentKey={log.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        {log.requestKey && (
          <li>
            <AttributeBadge name='request' value={log.requestKey} />
          </li>
        )}
        {log.error && (
          <li>
            <AttributeBadge name='error' value={log.error} />
          </li>
        )}
      </ul>
    </div>
  )
}
