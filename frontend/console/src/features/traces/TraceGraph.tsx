import { CallEvent, type Event, IngressEvent, PubSubConsumeEvent, PubSubPublishEvent } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { HoverPopup } from '../../shared/components/HoverPopup'
import { classNames, durationToMillis } from '../../shared/utils'
import { type TraceEvent, useRequestTraceEvents } from '../timeline/hooks/use-request-trace-events'
import { eventBackgroundColor } from '../timeline/timeline.utils'
import { eventBarLeftOffsetPercentage, requestStartTime, totalDurationForRequest } from './traces.utils'

const EventBlock = ({
  event,
  isSelected,
  requestStartTime,
  requestDuration,
}: {
  event: Event
  isSelected: boolean
  requestStartTime: number
  requestDuration: number
}) => {
  const traceEvent = event.entry.value as TraceEvent
  const durationInMillis = traceEvent.duration ? durationToMillis(traceEvent.duration) : 0
  let width = (durationInMillis / requestDuration) * 100
  if (width < 1) {
    width = 1
  }

  const leftOffsetPercentage = eventBarLeftOffsetPercentage(event, requestStartTime, requestDuration)

  let eventType = ''
  let eventTarget = ''

  if (traceEvent instanceof CallEvent) {
    eventType = 'call'
    eventTarget = `${traceEvent.destinationVerbRef?.module}.${traceEvent.destinationVerbRef?.name}`
  } else if (traceEvent instanceof IngressEvent) {
    eventType = 'ingress'
    eventTarget = traceEvent.path
  } else if (traceEvent instanceof PubSubPublishEvent) {
    eventType = 'publish'
    eventTarget = traceEvent.topic
  } else if (traceEvent instanceof PubSubConsumeEvent) {
    eventType = 'consume'
    eventTarget = traceEvent.topic
  }

  const barColor = isSelected ? 'bg-green-500' : eventBackgroundColor(event)

  const popupContent = (
    <p>
      {eventType} <span className='text-indigo-500 dark:text-indigo-400'>{eventTarget}</span>
      {` (${durationInMillis} ms)`}
    </p>
  )

  return (
    <div className='group relative my-0.5 h-2.5 flex'>
      <HoverPopup popupContent={popupContent} className='flex-grow relative'>
        <div
          className={classNames('absolute h-2.5 rounded-sm', barColor)}
          style={{
            width: `${width}%`,
            left: `${leftOffsetPercentage}%`,
          }}
        />
      </HoverPopup>
    </div>
  )
}

export const TraceGraph = ({ requestKey, selectedEventId }: { requestKey?: string; selectedEventId?: bigint }) => {
  const requestEvents = useRequestTraceEvents(requestKey)
  const events = requestEvents.data ?? []

  if (events.length === 0) {
    return
  }

  const startTime = requestStartTime(events)
  const totalEventDuration = totalDurationForRequest(events)

  return (
    <div className='flex flex-col'>
      {events
        .sort((a, b) => requestStartTime([a]) - requestStartTime([b]))
        .map((c, index) => (
          <div key={index} className='flex hover:bg-indigo-500/60 hover:dark:bg-slate-700 rounded-sm'>
            <div className='w-full relative'>
              <EventBlock event={c} isSelected={c.id === selectedEventId} requestStartTime={startTime} requestDuration={totalEventDuration} />
            </div>
          </div>
        ))}
    </div>
  )
}
