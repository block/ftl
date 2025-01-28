import type { NodeProps } from 'reactflow'

interface Props extends NodeProps {
  data: {
    title: string
    selected: boolean
  }
}

export const GroupNode = ({ data }: Props) => {
  return (
    <>
      <div className={`h-full rounded-md ${data.selected ? 'ring-2 ring-pink-400 dark:ring-pink-600/80' : 'ring-2 ring-indigo-400 dark:ring-indigo-800/80'}`}>
        <div
          className={`absolute rounded-t-md top-0 left-0 right-0 h-[30px] flex items-center px-4 ${
            data.selected ? 'bg-pink-400/20 dark:bg-pink-600/80' : 'bg-indigo-100 dark:bg-indigo-800/80'
          }`}
        >
          <div className='text-xs font-medium dark:text-gray-100 truncate'>{data.title}</div>
        </div>
      </div>
    </>
  )
}
