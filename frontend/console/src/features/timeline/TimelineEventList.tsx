import type { Event } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { HoverPopup } from '../../shared/components/HoverPopup'
import { formatTimestampShort } from '../../shared/utils'
import { deploymentTextColor } from '../deployments/deployment.utils'
import { TimelineAsyncExecute } from './TimelineAsyncExecute'
import { TimelineCall } from './TimelineCall'
import { TimelineChangesetChanged } from './TimelineChangesetChanged'
import { TimelineChangesetCreated } from './TimelineChangesetCreated'
import { TimelineCronScheduled } from './TimelineCronScheduled'
import { TimelineDeploymentCreated } from './TimelineDeploymentCreated'
import { TimelineIcon } from './TimelineIcon'
import { TimelineIngress } from './TimelineIngress'
import { TimelineLog } from './TimelineLog'
import { TimelinePubSubConsume } from './TimelinePubSubConsume'
import { TimelinePubSubPublish } from './TimelinePubSubPublish'
import { TimelineRuntime } from './TimelineRuntime'

interface EventTimelineProps {
  events: Event[]
  selectedEventId?: bigint
  handleEntryClicked: (entry: Event) => void
}

const deploymentKey = (event: Event) => {
  switch (event.entry?.case) {
    case 'call':
    case 'log':
    case 'ingress':
    case 'cronScheduled':
    case 'asyncExecute':
    case 'pubsubConsume':
    case 'pubsubPublish':
      return event.entry.value.deploymentKey
    case 'changesetCreated':
    case 'changesetStateChanged':
      return event.entry.value.key
    case 'deploymentCreated':
    case 'deploymentRuntime':
      return event.entry.value.key
    default:
      return ''
  }
}

const moduleName = (event: Event) => {
  if (event.entry?.case === 'changesetCreated' || event.entry?.case === 'changesetStateChanged') {
    return 'changeset'
  }
  const key = deploymentKey(event)
  const parts = key.split('-')
  return parts.length >= 2 ? parts[1] : ''
}

const DeploymentCell = ({ entry }: { entry: Event }) => {
  return (
    <td className={`p-1 pr-2 w-24 items-center flex-none truncate ${deploymentTextColor(deploymentKey(entry))}`}>
      <HoverPopup popupContent={deploymentKey(entry)}>{moduleName(entry)}</HoverPopup>
    </td>
  )
}

export const TimelineEventList = ({ events, selectedEventId, handleEntryClicked }: EventTimelineProps) => {
  return (
    <div className='overflow-x-hidden'>
      <table className={'w-full table-fixed text-gray-600 dark:text-gray-300'}>
        <thead>
          <tr className='flex text-xs'>
            <th className='p-1 text-left border-b w-8 border-gray-100 dark:border-slate-700 flex-none' />
            <th className='p-1 text-left border-b w-40 border-gray-100 dark:border-slate-700 flex-none'>Date</th>
            <th className='p-1 text-left border-b w-24 border-gray-100 dark:border-slate-700 flex-none'>Module</th>
            <th className='p-1 text-left border-b border-gray-100 dark:border-slate-700 flex-grow flex-shrink'>Content</th>
          </tr>
        </thead>
        <tbody>
          {events.map((entry) => (
            <tr
              key={entry.id}
              className={`flex border-b border-gray-100 dark:border-slate-700 text-xs font-roboto-mono ${
                selectedEventId === entry.id ? 'bg-indigo-50 dark:bg-slate-700' : ''
              } relative flex cursor-pointer hover:bg-indigo-50 dark:hover:bg-slate-700`}
              onClick={() => handleEntryClicked(entry)}
            >
              <td className='w-8 flex-none flex items-center justify-center'>
                <TimelineIcon event={entry} />
              </td>
              <td className='p-1 w-40 items-center flex-none text-gray-400 dark:text-gray-400'>{formatTimestampShort(entry.timestamp)}</td>
              <DeploymentCell entry={entry} />
              <td className='p-1 flex-grow truncate'>
                {(() => {
                  switch (entry.entry.case) {
                    case 'call':
                      return <TimelineCall call={entry.entry.value} />
                    case 'log':
                      return <TimelineLog log={entry.entry.value} />
                    case 'ingress':
                      return <TimelineIngress ingress={entry.entry.value} />
                    case 'cronScheduled':
                      return <TimelineCronScheduled cron={entry.entry.value} />
                    case 'asyncExecute':
                      return <TimelineAsyncExecute asyncExecute={entry.entry.value} />
                    case 'pubsubPublish':
                      return <TimelinePubSubPublish pubSubPublish={entry.entry.value} />
                    case 'pubsubConsume':
                      return <TimelinePubSubConsume pubSubConsume={entry.entry.value} />
                    case 'changesetCreated':
                      return <TimelineChangesetCreated changeset={entry.entry.value} />
                    case 'changesetStateChanged':
                      return <TimelineChangesetChanged changeset={entry.entry.value} />
                    case 'deploymentCreated':
                      return <TimelineDeploymentCreated deployment={entry.entry.value} />
                    case 'deploymentRuntime':
                      return <TimelineRuntime runtime={entry.entry.value} />
                    default:
                      return null
                  }
                })()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default TimelineEventList
