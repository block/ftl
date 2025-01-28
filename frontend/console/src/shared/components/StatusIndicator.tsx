interface StatusIndicatorProps {
  state: 'success' | 'error' | 'new' | 'busy' | 'idle'
  text?: string
}

export const StatusIndicator = ({ state, text }: StatusIndicatorProps) => {
  const stateClasses = {
    success: 'bg-green-400',
    error: 'bg-red-400',
    new: 'bg-blue-400',
    busy: 'bg-blue-400',
    idle: 'bg-gray-400',
  }

  return (
    <div className='flex items-center gap-x-2'>
      <div className={`h-2 w-2 rounded-full ${stateClasses[state]}`} />
      {text && <span>{text}</span>}
    </div>
  )
}
