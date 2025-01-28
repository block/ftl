import { Activity03Icon } from 'hugeicons-react'
import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '../../shared/components/Button'
import { useRequestTraceEvents } from '../timeline/hooks/use-request-trace-events'
import { totalDurationForRequest } from './traces.utils'

export const TraceGraphHeader = ({ requestKey, eventId }: { requestKey?: string; eventId: bigint }) => {
  const navigate = useNavigate()
  const requestEvents = useRequestTraceEvents(requestKey)
  const events = requestEvents.data ?? []

  const totalEventDuration = useMemo(() => totalDurationForRequest(events), [events])

  if (events.length === 0) {
    return null
  }

  return (
    <div className='flex items-center justify-between mb-1'>
      <span className='text-xs font-mono'>
        Total <span>{totalEventDuration.toFixed(2)}ms</span>
      </span>

      <Button variant='secondary' size='sm' onClick={() => navigate(`/traces/${requestKey}?event_id=${eventId}`)} title='View trace'>
        <Activity03Icon className='size-5' />
      </Button>
    </div>
  )
}
