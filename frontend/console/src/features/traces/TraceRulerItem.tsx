import { TraceGraphRuler } from './TraceGraphRuler'

export const TraceRulerItem = ({ duration }: { duration: number }) => {
  return (
    <div className='w-full flex items-center'>
      {/* Left spacer to align with trace items */}
      <div className='w-1/3 flex-shrink-0' />

      {/* Ruler container - match exact width of the bars container */}
      <div className='w-2/3 flex-grow pr-24'>
        <TraceGraphRuler duration={duration} />
      </div>
    </div>
  )
}
