import {
  Call02Icon,
  CallIncoming04Icon,
  ComputerIcon,
  Download04Icon,
  InternetIcon,
  KeyframesMultipleAddIcon,
  KeyframesMultipleIcon,
  Menu01Icon,
  TimeQuarterPassIcon,
  Upload04Icon,
} from 'hugeicons-react'
import type { Event } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { LogLevelBadgeSmall } from '../logs/LogLevelBadgeSmall'
import { eventTextColor } from './timeline.utils'

export const TimelineIcon = ({ event }: { event: Event }) => {
  const icon = (event: Event) => {
    const style = 'h4 w-4'
    const textColor = eventTextColor(event)

    switch (event.entry.case) {
      case 'call': {
        return event.entry.value.sourceVerbRef ? <Call02Icon className={`${style} ${textColor}`} /> : <CallIncoming04Icon className={`${style} ${textColor}`} />
      }
      case 'log':
        return <LogLevelBadgeSmall logLevel={event.entry.value.logLevel} />
      case 'ingress':
        return <InternetIcon className={`${style} ${textColor}`} />
      case 'cronScheduled':
        return <TimeQuarterPassIcon className={`${style} ${textColor}`} />
      case 'pubsubPublish':
        return <Upload04Icon className={`${style} ${textColor}`} />
      case 'pubsubConsume':
        return <Download04Icon className={`${style} ${textColor}`} />
      case 'changesetCreated':
        return <KeyframesMultipleAddIcon className={`${style} ${textColor}`} />
      case 'changesetStateChanged':
        return <KeyframesMultipleIcon className={`${style} ${textColor}`} />
      case 'deploymentRuntime':
        return <ComputerIcon className={`${style} ${textColor}`} />
      default:
        return <Menu01Icon className={`${style}`} />
    }
  }

  return <div>{icon(event)}</div>
}
