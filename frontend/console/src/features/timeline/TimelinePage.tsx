import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import type { TimelineQuery_Filter } from '../../protos/xyz/block/ftl/timeline/v1/timeline_pb'
import { SidePanelProvider } from '../../shared/providers/side-panel-provider.tsx'
import { Timeline } from './Timeline'
import { TimelineFilterPanel } from './filters/TimelineFilterPanel'
import { TIME_RANGES, type TimeSettings, TimelineTimeControls } from './filters/TimelineTimeControls'

export const TimelinePage = () => {
  const [searchParams] = useSearchParams()
  const [timeSettings, setTimeSettings] = useState<TimeSettings>({ isTailing: true, isPaused: false })
  const [filters, setFilters] = useState<TimelineQuery_Filter[]>([])
  const [selectedTimeRange, setSelectedTimeRange] = useState(TIME_RANGES.tail)
  const [isTimelinePaused, setIsTimelinePaused] = useState(false)

  const initialEventId = searchParams.get('id')
  useEffect(() => {
    if (initialEventId) {
      // if we're loading a specific event, we don't want to tail.
      setSelectedTimeRange(TIME_RANGES['24h'])
      setIsTimelinePaused(true)
    }
  }, [])

  const handleTimeSettingsChanged = (settings: TimeSettings) => {
    setTimeSettings(settings)
  }

  const handleFiltersChanged = (filters: TimelineQuery_Filter[]) => {
    setFilters(filters)
  }

  return (
    <SidePanelProvider>
      <div className='flex h-full'>
        <div className='flex-none overflow-y-auto'>
          <TimelineTimeControls selectedTimeRange={selectedTimeRange} isTimelinePaused={isTimelinePaused} onTimeSettingsChange={handleTimeSettingsChanged} />
          <TimelineFilterPanel onFiltersChanged={handleFiltersChanged} />
        </div>
        <div className='flex-grow overflow-y-scroll'>
          <Timeline timeSettings={timeSettings} filters={filters} />
        </div>
      </div>
    </SidePanelProvider>
  )
}
