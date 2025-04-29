import type { EngineEvent } from '../../protos/xyz/block/ftl/buildengine/v1/buildengine_pb'

type EventCase = EngineEvent['event']['case']

interface BuildStatusIndicatorProps {
  eventCase: EventCase | undefined
  text?: string
}

export const BuildStatusIndicator = ({ eventCase, text }: BuildStatusIndicatorProps) => {
  let bgColor: string | undefined
  let isBusy = false

  switch (eventCase) {
    case 'moduleBuildWaiting':
      bgColor = 'bg-gray-400'
      isBusy = true
      break
    case 'moduleDeployWaiting':
      bgColor = 'bg-gray-400'
      isBusy = true
      break
    case 'moduleBuildStarted':
      bgColor = 'bg-blue-400'
      isBusy = true
      break
    case 'moduleDeployStarted':
      bgColor = 'bg-green-400'
      isBusy = true
      break
    case 'moduleBuildSuccess':
      bgColor = 'bg-blue-400' // Built
      break
    case 'moduleDeploySuccess':
      bgColor = 'bg-green-400' // Deployed
      break
    case 'moduleBuildFailed':
    case 'moduleDeployFailed':
      bgColor = 'bg-red-400' // Failed
      break
  }

  return (
    <div className='flex items-center gap-x-2'>
      {bgColor && (
        <span className='relative flex h-2 w-2'>
          {isBusy && <span className={`absolute inline-flex h-full w-full animate-ping rounded-full opacity-75 ${bgColor}`} />}
          <span className={`relative inline-flex h-2 w-2 rounded-full ${bgColor}`} />
        </span>
      )}
      {text && <span>{text}</span>}
    </div>
  )
}
