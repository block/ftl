import { Activity03Icon, ArrowLeft02Icon } from 'hugeicons-react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { Button } from '../../shared/components/Button'
import { Divider } from '../../shared/components/Divider'
import { Loader } from '../../shared/components/Loader'
import { useRequestTraceEvents } from '../timeline/hooks/use-request-trace-events'
import { TraceDetails } from './TraceDetails'
import { TraceDetailsCall } from './details/TraceDetailsCall'
import { TraceDetailsIngress } from './details/TraceDetailsIngress'
import { TraceDetailsPubsubConsume } from './details/TraceDetailsPubsubConsume'
import { TraceDetailsPubsubPublish } from './details/TraceDetailsPubsubPublish'

const RequestNotFound = ({ requestKey, onBack }: { requestKey: string; onBack: () => void }) => (
  <div className='flex flex-col items-center justify-center min-h-screen'>
    <Activity03Icon className='size-16 text-gray-400 dark:text-gray-500 mb-4' />
    <h2 className='text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2'>Request Not Found</h2>
    <p className='text-gray-600 dark:text-gray-400 mb-4'>No trace data found for request ID: {requestKey}</p>
    <Button variant='secondary' size='sm' onClick={onBack}>
      Go Back
    </Button>
  </div>
)

export const TracesPage = () => {
  const navigate = useNavigate()

  const { requestKey } = useParams<{ requestKey: string }>()
  const requestEvents = useRequestTraceEvents(requestKey)
  const events = requestEvents.data ?? []

  const [searchParams] = useSearchParams()
  const eventIdParam = searchParams.get('event_id')
  const selectedEventId = eventIdParam ? BigInt(eventIdParam) : undefined

  const handleBack = () => {
    if (window.history.length > 1) {
      navigate(-1)
    } else {
      navigate('/modules')
    }
  }

  if (requestKey === undefined) {
    return <RequestNotFound requestKey='Invalid Request ID' onBack={handleBack} />
  }

  if (requestEvents.isLoading) {
    return (
      <div className='flex justify-center items-center min-h-screen'>
        <Loader />
      </div>
    )
  }

  if (!requestEvents.isLoading && events.length === 0) {
    return <RequestNotFound requestKey={requestKey} onBack={handleBack} />
  }

  if (!selectedEventId && events.length > 0) {
    const firstEventId = events[0].id
    navigate(`/traces/${requestKey}?event_id=${firstEventId}`, { replace: true })
    return (
      <div className='flex justify-center items-center min-h-screen'>
        <Loader />
      </div>
    )
  }

  const selectedEvent = events.find((event) => event.id === selectedEventId)
  let eventDetailsComponent: React.ReactNode
  switch (selectedEvent?.entry.case) {
    case 'call':
      eventDetailsComponent = <TraceDetailsCall event={selectedEvent} />
      break
    case 'ingress':
      eventDetailsComponent = <TraceDetailsIngress event={selectedEvent} />
      break
    case 'pubsubPublish':
      eventDetailsComponent = <TraceDetailsPubsubPublish event={selectedEvent} />
      break
    case 'pubsubConsume':
      eventDetailsComponent = <TraceDetailsPubsubConsume event={selectedEvent} />
      break
    default:
      eventDetailsComponent = <p>No details available for this event type.</p>
      break
  }

  return (
    <div className='flex h-full'>
      <div className='w-1/2 p-4 h-full overflow-y-auto'>
        <div className='flex items-center mb-2'>
          <Button variant='secondary' size='sm' onClick={handleBack} title='Back'>
            <ArrowLeft02Icon className='size-6' />
          </Button>
          <span className='text-xl font-semibold ml-2'>Trace Details</span>
        </div>
        <TraceDetails requestKey={requestKey} events={events} selectedEventId={selectedEventId} />
      </div>

      <Divider vertical />

      <div className='w-1/2 p-4 mt-1 h-full overflow-y-auto'>{eventDetailsComponent}</div>
    </div>
  )
}
