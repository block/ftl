import { Handle, type NodeProps, Position } from '@xyflow/react'
import { ViewOffSlashIcon } from 'hugeicons-react'
import { Verb } from '../../protos/xyz/block/ftl/schema/v1/schema_pb'
import { declIcon } from '../modules/module.utils'

interface Props extends NodeProps {
  data: {
    title: string
    selected: boolean
    nodeType?: string
    style?: {
      backgroundColor?: string
    }
    id: string
    isExported?: boolean
  }
}

export const DeclNode = ({ data }: Props) => {
  const handleColor = data.selected ? 'rgb(251 113 133)' : data.style?.backgroundColor || 'rgb(79 70 229)'
  const defaultVerb = new Verb({ name: data.title })
  const Icon = declIcon(data.nodeType || 'verb', defaultVerb)
  const isPrivate = data.isExported === false

  return (
    <div className={`rounded-md overflow-hidden ${data.selected ? 'ring-2 ring-pink-400 dark:ring-pink-600' : ''}`}>
      <Handle id={`${data.id}-target`} type='target' position={Position.Left} style={{ border: 0, backgroundColor: handleColor }} isConnectable={false} />

      <div className='flex' style={{ backgroundColor: data.style?.backgroundColor }}>
        <div className='flex items-center text-gray-100 px-3 py-2 gap-2 w-full'>
          <Icon className='size-4 flex-shrink-0' />
          <div className='text-xs truncate min-w-0 flex-1'>{data.title}</div>
          {isPrivate && <ViewOffSlashIcon className='size-3.5 flex-shrink-0 text-gray-300' aria-label='Private (not exported)' />}
        </div>
      </div>

      <Handle id={`${data.id}-source`} type='source' position={Position.Right} style={{ border: 0, backgroundColor: handleColor }} isConnectable={false} />
    </div>
  )
}
