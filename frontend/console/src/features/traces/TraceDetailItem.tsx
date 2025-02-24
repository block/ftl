import { AsyncExecuteEvent, CallEvent, type Event, IngressEvent, PubSubConsumeEvent, PubSubPublishEvent } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { HoverPopup } from '../../shared/components/HoverPopup'
import { classNames } from '../../shared/utils'
import { TimelineIcon } from '../timeline/TimelineIcon'
import type { TraceEvent } from '../timeline/hooks/use-request-trace-events'
import { eventBackgroundColor } from '../timeline/timeline.utils'
import { eventBarLeftOffsetPercentage } from './traces.utils'

interface TraceDetailItemProps {
  event: Event
  traceEvent: TraceEvent
  eventDurationMs: number
  requestDurationMs: number
  requestStartTime: number
  selectedEventId: bigint | undefined
  handleEventClick: (eventId: bigint) => void
}

export const TraceDetailItem: React.FC<TraceDetailItemProps> = ({
  event,
  traceEvent,
  eventDurationMs,
  requestDurationMs,
  requestStartTime,
  selectedEventId,
  handleEventClick,
}) => {
  const leftOffsetPercentage = eventBarLeftOffsetPercentage(event, requestStartTime, requestDurationMs)

  let width = (eventDurationMs / requestDurationMs) * 100
  if (width < 1) width = 1

  let action = ''
  let eventName = ''
  const icon = <TimelineIcon event={event} />

  if (traceEvent instanceof CallEvent) {
    action = 'Call'
    eventName = `${traceEvent.destinationVerbRef?.module}.${traceEvent.destinationVerbRef?.name}`
  } else if (traceEvent instanceof IngressEvent) {
    action = `HTTP ${traceEvent.method}`
    eventName = `${traceEvent.path}`
  } else if (traceEvent instanceof AsyncExecuteEvent) {
    action = 'Async'
    eventName = `${traceEvent.verbRef?.module}.${traceEvent.verbRef?.name}`
  } else if (traceEvent instanceof PubSubPublishEvent) {
    action = 'Publish'
    eventName = `${traceEvent.topic}`
  } else if (traceEvent instanceof PubSubConsumeEvent) {
    action = 'Consume'
    eventName = `${traceEvent.destVerbModule}.${traceEvent.destVerbName}`
  }

  const barColor = event.id === selectedEventId ? 'bg-green-500' : eventBackgroundColor(event)

  const isSelected = event.id === selectedEventId
  const listItemClass = classNames(
    'flex items-center w-full p-2 rounded cursor-pointer',
    isSelected ? 'bg-indigo-100/50 dark:bg-indigo-700/50' : 'hover:bg-indigo-500/10',
  )

  return (
    <li key={event.id.toString()} className={listItemClass} onClick={() => handleEventClick(event.id)}>
      <div className='w-1/3 flex-shrink-0 flex items-center gap-x-2 pr-2 font-medium text-sm overflow-hidden'>
        <HoverPopup popupContent={<span>{action}</span>}>
          <span className='flex-shrink-0'>{icon}</span>
        </HoverPopup>
        <span className='truncate'>{eventName}</span>
      </div>

      <div className='w-2/3 flex-grow relative h-5 pr-24'>
        <div
          className={`absolute h-5 ${barColor} rounded-sm`}
          style={{
            width: `${width}%`,
            left: `${leftOffsetPercentage}%`,
            maxWidth: '100%',
          }}
        />
      </div>

      <div className='w-24 flex-shrink-0 text-xs font-medium text-right'>{eventDurationMs.toFixed(2)} ms</div>
    </li>
  )
}
