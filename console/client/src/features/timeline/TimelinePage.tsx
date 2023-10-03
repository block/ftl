import { ListBulletIcon } from '@heroicons/react/24/outline'
import React from 'react'
import { useSearchParams } from 'react-router-dom'
import { Page } from '../../layout'
import { EventsQuery_Filter } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { Timeline } from './Timeline'
import { TimelineFilterPanel } from './filters/TimelineFilterPanel'
import { TIME_RANGES, TimeSettings, TimelineTimeControls } from './filters/TimelineTimeControls'

export const TimelinePage = () => {
  const [searchParams] = useSearchParams()
  const [timeSettings, setTimeSettings] = React.useState<TimeSettings>({ isTailing: true, isPaused: false })
  const [filters, setFilters] = React.useState<EventsQuery_Filter[]>([])
  const [selectedTimeRange, setSelectedTimeRange] = React.useState(TIME_RANGES['tail'])
  const [isTimelinePaused, setIsTimelinePaused] = React.useState(false)

  React.useEffect(() => {
    if (searchParams.get('id')) {
      // if we're loading a specific event, we don't want to tail.
      setSelectedTimeRange(TIME_RANGES['5m'])
      setIsTimelinePaused(true)
    }
  }, [searchParams])

  const handleTimeSettingsChanged = (settings: TimeSettings) => {
    setTimeSettings(settings)
  }

  const handleFiltersChanged = (filters: EventsQuery_Filter[]) => {
    setFilters(filters)
  }

  return (
    <Page>
    <Page.Header icon={<ListBulletIcon />} title='Events'>
      <TimelineTimeControls
        selectedTimeRange={selectedTimeRange}
        isTimelinePaused={isTimelinePaused}
        onTimeSettingsChange={handleTimeSettingsChanged}
      />
    </Page.Header>
    <Page.Body className='flex'>
      <div className='sticky top-0 flex-none overflow-y-auto'>
        <TimelineFilterPanel onFiltersChanged={handleFiltersChanged} />
      </div>
      <div className='flex-grow overflow-y-scroll'>
        <Timeline timeSettings={timeSettings} filters={filters} />
      </div>
    </Page.Body>
  </Page>
  )
}
