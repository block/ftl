import { CodeBlock } from './CodeBlock'

interface Props {
  title: string
  code: string
  maxHeight?: number
}

export const CodeBlockWithTitle = ({ title, code, maxHeight }: Props) => {
  const formattedCode = (() => {
    try {
      // If it's already a JSON string, parse and re-stringify to ensure consistent formatting
      const parsed = JSON.parse(code)
      return JSON.stringify(parsed, null, 2)
    } catch {
      // If it's not valid JSON, return as-is
      return code
    }
  })()

  return (
    <div className='mt-4'>
      <div className='text-xs font-medium text-gray-600 dark:text-gray-400 mb-1 uppercase tracking-wider'>{title}</div>
      <CodeBlock code={formattedCode} maxHeight={maxHeight} />
    </div>
  )
}
