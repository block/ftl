export interface AttributeBadgeProps {
  name: string
  value: string
}

export const AttributeBadge = ({ name, value, ...props }: AttributeBadgeProps) => {
  return (
    <div className='inline-flex rounded-md text-xs font-medium h-6' {...props}>
      <span className='px-2 flex items-center text-gray-400 bg-gray-100 border border-gray-200 dark:border-gray-600 rounded-s-md dark:bg-gray-700'>{name}</span>
      <span className='px-2 flex items-center text-gray-900 border-t border-b border-r border-gray-200 dark:border-gray-600 rounded-r-md dark:bg-gray-800 dark:text-white max-w-[200px] truncate'>
        {value}
      </span>
    </div>
  )
}
