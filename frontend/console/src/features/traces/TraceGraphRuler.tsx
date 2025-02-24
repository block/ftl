import { useEffect, useRef, useState } from 'react'

export const TraceGraphRuler = ({ duration }: { duration: number }) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const [tickCount, setTickCount] = useState(5)

  useEffect(() => {
    const updateTickCount = () => {
      if (!containerRef.current) return

      const width = containerRef.current.offsetWidth
      // Use fewer ticks for smaller widths
      if (width < 200) {
        setTickCount(3) // 0ms, 12ms, 24ms
      } else {
        setTickCount(5) // 0ms, 6ms, 12ms, 18ms, 24ms
      }
    }

    // Initial update
    updateTickCount()

    // Update on resize
    const resizeObserver = new ResizeObserver(updateTickCount)
    if (containerRef.current) {
      resizeObserver.observe(containerRef.current)
    }

    return () => {
      if (containerRef.current) {
        resizeObserver.unobserve(containerRef.current)
      }
    }
  }, [])

  // Calculate ticks based on current tick count
  const tickInterval = duration / (tickCount - 1)
  const ticks = Array.from({ length: tickCount }, (_, i) => ({
    value: Math.round(i * tickInterval),
    position: `${(i * 100) / (tickCount - 1)}%`,
  }))

  return (
    <div ref={containerRef} className='relative w-full h-8 mb-2'>
      <div className='absolute bottom-0 left-0 right-0 h-[1px] bg-gray-200 dark:bg-gray-600' />

      {ticks.map((tick, index) => {
        const isLastTick = index === tickCount - 1

        return (
          <div
            key={index}
            className='absolute bottom-0 transform -translate-x-1/2'
            style={{
              left: isLastTick ? 'calc(100% )' : tick.position,
            }}
          >
            <span className='absolute bottom-3 text-xs font-roboto-mono text-gray-500 dark:text-gray-400 -translate-x-1/2 whitespace-nowrap'>
              {tick.value}ms
            </span>
            <span className='block h-2 w-[1px] bg-gray-200 dark:bg-gray-600' />
          </div>
        )
      })}
    </div>
  )
}
