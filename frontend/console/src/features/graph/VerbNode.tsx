import { CodeCircleIcon } from 'hugeicons-react'
import { Handle, type NodeProps, Position } from 'reactflow'

export const verbHeight = 40

interface Props extends NodeProps {
  data: {
    title: string
    selected: boolean
  }
}

export const VerbNode = ({ data }: Props) => {
  const handleColor = data.selected ? 'rgb(251 113 133)' : 'rgb(79 70 229)'
  return (
    <>
      <Handle type='target' position={Position.Left} style={{ border: 0, backgroundColor: handleColor }} isConnectable={true} />

      <div className={`flex h-full w-full bg-indigo-600 rounded-md ${data.selected ? 'bg-pink-600' : ''}`}>
        <div className='flex items-center text-gray-200 px-2 gap-1 w-full overflow-hidden'>
          <CodeCircleIcon className='size-4 flex-shrink-0' />
          <div className='text-xs truncate min-w-0 flex-1'>{data.title}</div>
        </div>
      </div>

      <Handle type='source' position={Position.Right} style={{ border: 0, backgroundColor: handleColor }} isConnectable={true} />
    </>
  )
}
