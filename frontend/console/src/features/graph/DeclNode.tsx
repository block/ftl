import {
  BubbleChatIcon,
  CodeCircleIcon,
  DatabaseIcon,
  FunctionIcon,
  LeftToRightListNumberIcon,
  MessageIncoming02Icon,
  Settings02Icon,
  SquareLock02Icon,
} from 'hugeicons-react'
import { Handle, type NodeProps, Position } from 'reactflow'

export const verbHeight = 40

interface Props extends NodeProps {
  data: {
    title: string
    selected: boolean
    nodeType?: string
  }
  style?: {
    backgroundColor?: string
  }
}

const getNodeIcon = (nodeType = 'verb') => {
  const icons = {
    verb: FunctionIcon,
    topic: BubbleChatIcon,
    database: DatabaseIcon,
    config: Settings02Icon,
    secret: SquareLock02Icon,
    enum: LeftToRightListNumberIcon,
    subscription: MessageIncoming02Icon,
    default: CodeCircleIcon,
  }
  return icons[nodeType as keyof typeof icons] || icons.default
}

export const DeclNode = ({ data, style }: Props) => {
  const handleColor = data.selected ? 'rgb(251 113 133)' : 'rgb(79 70 229)'
  const Icon = getNodeIcon(data.nodeType)

  return (
    <>
      <Handle type='target' position={Position.Left} style={{ border: 0, backgroundColor: handleColor }} isConnectable={true} />

      <div
        className='flex h-full w-full rounded-xl overflow-hidden'
        style={{
          background: style?.backgroundColor,
        }}
      >
        <div className='flex items-center text-gray-100 px-4 py-2 gap-2 w-full'>
          <Icon className='size-4 flex-shrink-0' />
          <div className='text-xs truncate min-w-0 flex-1'>{data.title}</div>
        </div>
      </div>

      <Handle type='source' position={Position.Right} style={{ border: 0, backgroundColor: handleColor }} isConnectable={true} />
    </>
  )
}
