import type { CronScheduledEvent, Event } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { AttributeBadge } from '../../../shared/components/AttributeBadge'
import { CodeBlockWithTitle } from '../../../shared/components/CodeBlockWithTitle'
import { formatDuration, formatTimestampShort } from '../../../shared/utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../modules/decls/verb/verb.utils'

export const TimelineCronScheduledDetails = ({ event }: { event: Event }) => {
  const cron = event.entry.value as CronScheduledEvent

  return (
    <>
      <div className='p-4'>
        {cron.error && <CodeBlockWithTitle title='Error' code={cron.error} />}
        <DeploymentCard deploymentKey={cron.deploymentKey} />

        <ul className='pt-4 space-y-2'>
          <li>
            <AttributeBadge name='duration' value={formatDuration(cron.duration)} />
          </li>
          {cron.verbRef && (
            <li>
              <AttributeBadge name='destination' value={refString(cron.verbRef)} />
            </li>
          )}
          {cron.schedule && (
            <li>
              <AttributeBadge name='schedule' value={cron.schedule} />
            </li>
          )}
          {cron.scheduledAt && (
            <li>
              <AttributeBadge name='scheduled for' value={formatTimestampShort(cron.scheduledAt)} />
            </li>
          )}
        </ul>
      </div>
    </>
  )
}
