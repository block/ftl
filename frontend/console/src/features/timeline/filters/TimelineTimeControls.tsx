import { Timestamp } from '@bufbuild/protobuf'
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from '@headlessui/react'
import { ArrowDown01Icon, Backward02Icon, Forward02Icon, PauseIcon, PlayIcon, Tick02Icon } from 'hugeicons-react'
import { useEffect, useState } from 'react'
import { Button } from '../../../shared/components/Button'
import { bgColor, borderColor, classNames, formatTimestampShort, formatTimestampTime, panelColor, textColor } from '../../../shared/utils'

interface TimeRange {
  label: string
  value: number
}

export const TIME_RANGES: Record<string, TimeRange> = {
  tail: { label: 'Live tail', value: 0 },
  '5m': { label: 'Past 5 minutes', value: 5 * 60 * 1000 },
  '15m': { label: 'Past 15 minutes', value: 15 * 60 * 1000 },
  '30m': { label: 'Past 30 minutes', value: 30 * 60 * 1000 },
  '1h': { label: 'Past 1 hour', value: 60 * 60 * 1000 },
  '24h': { label: 'Past 24 hours', value: 24 * 60 * 60 * 1000 },
}

export interface TimeSettings {
  isTailing: boolean
  isPaused: boolean
  olderThan?: Timestamp
  newerThan?: Timestamp
}

export const TimelineTimeControls = ({
  onTimeSettingsChange,
  selectedTimeRange,
  isTimelinePaused,
}: {
  onTimeSettingsChange: (settings: TimeSettings) => void
  selectedTimeRange: TimeRange
  isTimelinePaused: boolean
}) => {
  const [selected, setSelected] = useState(selectedTimeRange)
  const [isPaused, setIsPaused] = useState(isTimelinePaused)
  const [newerThan, setNewerThan] = useState<Timestamp | undefined>()

  const isTailing = selected.value === TIME_RANGES.tail.value

  useEffect(() => {
    handleRangeChanged(selectedTimeRange)
    setIsPaused(isTimelinePaused)
  }, [selectedTimeRange, isTimelinePaused])

  useEffect(() => {
    if (isTailing) {
      onTimeSettingsChange({ isTailing, isPaused })
      return
    }

    if (newerThan) {
      const startTime = (newerThan.toDate() ?? new Date()).getTime()
      const olderThanDate = new Date(startTime + selected.value)

      onTimeSettingsChange({
        isTailing,
        isPaused,
        olderThan: Timestamp.fromDate(olderThanDate),
        newerThan: newerThan,
      })
    }
  }, [selected, isPaused, newerThan])

  const handleRangeChanged = (range: TimeRange) => {
    setSelected(range)

    if (range.value === TIME_RANGES.tail.value) {
      setNewerThan(undefined)
      setIsPaused(false)
    } else {
      const newerThanDate = new Date(new Date().getTime() - range.value)
      setNewerThan(Timestamp.fromDate(newerThanDate))
    }
  }

  const handleTimeBackward = () => {
    if (!newerThan) {
      return
    }
    const newerThanDate = new Date(newerThan.toDate().getTime() - selected.value)
    setNewerThan(Timestamp.fromDate(newerThanDate))
  }

  const handleTimeForward = () => {
    if (!newerThan) {
      return
    }
    const newerThanTime = newerThan.toDate().getTime()
    const newerThanDate = new Date(newerThanTime + selected.value)
    const maxNewTime = new Date().getTime() - selected.value
    if (newerThanDate.getTime() > maxNewTime) {
      setNewerThan(Timestamp.fromDate(new Date(maxNewTime)))
    } else {
      setNewerThan(Timestamp.fromDate(newerThanDate))
    }
  }

  const olderThan = newerThan ? Timestamp.fromDate(new Date(newerThan.toDate().getTime() - selected.value)) : undefined
  return (
    <>
      {newerThan && (
        <div
          title={`${formatTimestampShort(olderThan)} - ${formatTimestampShort(newerThan)}`}
          className='text-xs font-roboto-mono mr-2 text-gray-400 pl-4 mt-2'
        >
          {formatTimestampTime(olderThan)} - {formatTimestampTime(newerThan)}
        </div>
      )}
      <div className='flex items-center h-6 pl-2 mt-2'>
        <Listbox value={selected} onChange={handleRangeChanged}>
          <div className='relative w-full mr-1 mt-0.5 items-center'>
            <ListboxButton
              className={`relative w-full cursor-pointer rounded-md ${bgColor} ${textColor} py-1 pl-2 pr-10 text-xs text-left shadow-sm ring-1 ring-inset ${borderColor} focus:outline-none focus:ring-2 focus:ring-indigo-600`}
            >
              <span className='block truncate'>{selected.label}</span>
              <span className='pointer-events-none absolute inset-y-0 right-0 flex items-center pr-1'>
                <ArrowDown01Icon className='h-5 w-5 text-gray-400' aria-hidden='true' />
              </span>
            </ListboxButton>

            <ListboxOptions
              anchor='bottom'
              transition
              className={`z-10 mt-1 w-40 rounded-md ${panelColor} py-1 text-xs shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none`}
            >
              {Object.keys(TIME_RANGES).map((key) => {
                const timeRange = TIME_RANGES[key]
                return (
                  <ListboxOption
                    key={key}
                    className={({ focus }) => classNames(focus ? 'bg-indigo-600 text-white' : textColor, 'relative cursor-pointer select-none py-2 pl-3 pr-9')}
                    value={timeRange}
                  >
                    {({ selected, focus }) => (
                      <>
                        <span className={classNames(selected ? 'font-semibold' : 'font-normal', 'block truncate')}>{timeRange.label}</span>

                        {selected ? (
                          <span className={classNames(focus ? 'text-white' : 'text-indigo-600', 'absolute inset-y-0 right-0 flex items-center pr-2')}>
                            <Tick02Icon className='h-4 w-4' aria-hidden='true' />
                          </span>
                        ) : null}
                      </>
                    )}
                  </ListboxOption>
                )
              })}
            </ListboxOptions>
          </div>
        </Listbox>
        {isTailing && (
          <Button variant='secondary' size='xs' onClick={() => setIsPaused(!isPaused)} title={isPaused ? 'Resume' : 'Pause'}>
            {isPaused ? <PlayIcon className='w-4 h-4' /> : <PauseIcon className='w-4 h-4' />}
          </Button>
        )}
        {!isTailing && (
          <span className={`isolate inline-flex rounded-md shadow-sm h-6 ${textColor} ${bgColor}`}>
            <Button variant='secondary' size='xs' onClick={handleTimeBackward} className='rounded-r-none'>
              <Backward02Icon className='w-4 h-4' />
            </Button>
            <Button variant='secondary' size='xs' onClick={handleTimeForward} className='rounded-l-none'>
              <Forward02Icon className='w-4 h-4' />
            </Button>
          </span>
        )}
      </div>
    </>
  )
}
