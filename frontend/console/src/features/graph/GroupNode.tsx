import type { NodeProps } from 'reactflow'

export const groupPadding = 40

interface Props extends NodeProps {
  data: {
    title: string
    selected: boolean
  }
}

export const GroupNode = ({ data }: Props) => {
  return (
    <>
      <div className={`h-full rounded-md ${data.selected ? 'bg-pink-400/80 dark:bg-pink-600/80' : 'bg-indigo-400/30 dark:bg-indigo-900/30'}`}>
        <div className='absolute top-0 left-0 right-0 h-[30px] flex items-center px-4'>
          <div className='text-xs font-medium dark:text-gray-100 truncate'>{data.title}</div>
        </div>
      </div>
    </>
  )
}
