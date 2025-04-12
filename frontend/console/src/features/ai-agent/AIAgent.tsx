import { ChevronDown, ChevronUp, GripVertical, Send as SendIcon, X as XIcon } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import type { Components } from 'react-markdown'
import ReactMarkdown from 'react-markdown'
import { classNames } from '../../shared/utils'
import { GooseIcon } from './GooseIcon'
import { useGoose } from './use-goose'

type Message = {
  type: 'user' | 'assistant'
  content: string
  isCode?: boolean
  isDiff?: boolean
  language?: string
}

type CodeBlockProps = {
  content: string
  isDiff: boolean
  previewLines?: number
  language?: string
}

type CodeBlockState = {
  isActive: boolean
  type: 'diff' | 'regular' | null
  language?: string
}

// Strip ANSI codes but preserve colors by converting them to CSS classes
const processAnsiLine = (line: string): { text: string; className: string } => {
  // Basic ANSI color code mapping to Tailwind classes
  const colorMap: Record<string, string> = {
    '[32m': 'text-green-400', // Green
    '[31m': 'text-red-400', // Red
    '[0m': '', // Reset
    '[1;38;2;222;222;222m': 'text-gray-300', // Light gray
    '[38;2;222;222;222m': 'text-gray-300', // Light gray
    '[38;2;254;214;175m': 'text-amber-200', // Amber for JSON
  }

  let className = ''
  // Remove ANSI codes but track the last color
  const escapeCode = String.fromCharCode(27)
  const cleanText = line.replace(new RegExp(`${escapeCode}\\[[0-9;:]*[mK]`, 'g'), (match) => {
    if (colorMap[match.slice(1)]) {
      className = colorMap[match.slice(1)]
    }
    return ''
  })

  // Preserve indentation by converting spaces to non-breaking spaces
  const preserveIndentation = cleanText.replace(/^(\s+)/, (match) => {
    return '\u00A0'.repeat(match.length)
  })

  return { text: preserveIndentation, className }
}

const CodeBlock = ({ content, isDiff, previewLines = 5, language }: CodeBlockProps) => {
  const [isExpanded, setIsExpanded] = useState(false)
  const lines = content.split('\n')
  const hasMoreLines = lines.length > previewLines

  // Extract and format the file header if it exists
  const fileHeaderMatch = stripAnsiCodes(lines[0]).match(/^###(.+)$/)
  const fileHeader = fileHeaderMatch ? fileHeaderMatch[1].trim() : null
  const displayLines = isExpanded ? (fileHeader ? lines.slice(1) : lines) : fileHeader ? lines.slice(1, previewLines + 1) : lines.slice(0, previewLines)

  return (
    <div className='relative w-full'>
      {fileHeader && (
        <div className='flex items-center gap-2 px-3 py-2 bg-gray-700 border-b border-gray-600 rounded-t text-sm font-mono'>
          <span className='text-blue-400 break-all'>{fileHeader}</span>
        </div>
      )}
      {language && (
        <div className='flex items-center gap-2 px-3 py-2 bg-gray-700 border-b border-gray-600 rounded-t text-sm font-mono'>
          <span className='text-blue-400'>{language}</span>
        </div>
      )}
      <div
        className={classNames(
          'overflow-x-auto overflow-y-hidden transition-all duration-200',
          !isExpanded && hasMoreLines && 'max-h-[250px]',
          (fileHeader || language) && 'rounded-t-none bg-gray-800',
          !fileHeader && !language && 'rounded-t bg-gray-800',
          'px-4 py-3',
        )}
      >
        {isDiff ? (
          <div className='diff-content font-mono min-w-full w-fit'>
            {displayLines.map((line, i) => {
              const processed = processAnsiLine(line)
              return (
                <div
                  key={i}
                  className={classNames(
                    'diff-line whitespace-pre break-keep',
                    processed.className,
                    line.startsWith('+') && 'text-green-400',
                    line.startsWith('-') && 'text-red-400',
                  )}
                >
                  {processed.text}
                </div>
              )
            })}
          </div>
        ) : (
          <div className='font-mono min-w-full w-fit'>
            {displayLines.map((line, i) => {
              const processed = processAnsiLine(line)
              return (
                <div key={i} className={classNames('whitespace-pre break-keep', processed.className)}>
                  {processed.text}
                </div>
              )
            })}
          </div>
        )}
      </div>
      {hasMoreLines && (
        <button
          type='button'
          onClick={() => setIsExpanded((prev) => !prev)}
          className='absolute bottom-0 left-0 right-0 py-2 px-2 flex items-center justify-center gap-1 text-xs text-gray-300 hover:text-white bg-gray-800 bg-opacity-90 rounded-b transition-colors border-t border-gray-700'
        >
          {isExpanded ? (
            <>
              <ChevronUp className='size-3' />
              <span>Show less</span>
            </>
          ) : (
            <>
              <ChevronDown className='size-3' />
              <span>Show more</span>
            </>
          )}
        </button>
      )}
    </div>
  )
}

function stripAnsiCodes(s: string): string {
  const escapeCode = String.fromCharCode(27)
  return s.replace(new RegExp(`${escapeCode}\\[[0-9;:]*[mK]`, 'g'), '')
}

const markdownComponents: Partial<Components> = {
  p: ({ children, ...props }) => (
    <p className='mb-4 last:mb-0 text-gray-100' {...props}>
      {children}
    </p>
  ),
  h1: ({ children, ...props }) => (
    <h1 className='text-2xl font-bold mb-4 text-white' {...props}>
      {children}
    </h1>
  ),
  h2: ({ children, ...props }) => (
    <h2 className='text-xl font-bold mb-3 text-white' {...props}>
      {children}
    </h2>
  ),
  h3: ({ children, ...props }) => (
    <h3 className='text-lg font-bold mb-2 text-white' {...props}>
      {children}
    </h3>
  ),
  ul: ({ children, ...props }) => (
    <ul className='list-disc pl-4 mb-4 last:mb-0 text-gray-100' {...props}>
      {children}
    </ul>
  ),
  ol: ({ children, ...props }) => (
    <ol className='list-decimal pl-4 mb-4 last:mb-0 text-gray-100' {...props}>
      {children}
    </ol>
  ),
  li: ({ children, ...props }) => (
    <li className='mb-1 last:mb-0' {...props}>
      {children}
    </li>
  ),
  code: ({ children, ...props }) => (
    <code className='bg-gray-800 rounded px-1 py-0.5 text-gray-100' {...props}>
      {children}
    </code>
  ),
  pre: ({ children, ...props }) => (
    <pre className='bg-gray-800 rounded p-3 overflow-x-auto mb-4' {...props}>
      {children}
    </pre>
  ),
  a: ({ children, href, ...props }) => (
    <a href={href} className='text-blue-400 hover:text-blue-300 underline' target='_blank' rel='noopener noreferrer' {...props}>
      {children}
    </a>
  ),
  blockquote: ({ children, ...props }) => (
    <blockquote className='border-l-4 border-gray-600 pl-4 italic mb-4 text-gray-300' {...props}>
      {children}
    </blockquote>
  ),
}

const MarkdownContent = ({ content }: { content: string }) => {
  return (
    <div className='prose prose-invert max-w-none'>
      <ReactMarkdown components={markdownComponents}>{content}</ReactMarkdown>
    </div>
  )
}

export const AIAgent = ({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) => {
  const [input, setInput] = useState('')
  const [messages, setMessages] = useState<Message[]>([])
  const [isProcessing, setIsProcessing] = useState(false)
  const [width, setWidth] = useState(33) // Width as percentage
  const [codeBlockState, setCodeBlockState] = useState<CodeBlockState>({
    isActive: false,
    type: null,
    language: undefined,
  })
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const panelRef = useRef<HTMLDivElement>(null)
  const isDraggingRef = useRef(false)
  const gooseMutation = useGoose()

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages, isProcessing])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isProcessing) return

    setIsProcessing(true)
    // Reset code block state at the start of a new message
    setCodeBlockState({
      isActive: false,
      type: null,
      language: undefined,
    })

    const userMessage = { type: 'user' as const, content: input }
    setMessages((prev) => [...prev, userMessage])
    const userInput = input
    setInput('')

    try {
      await gooseMutation.mutateAsync({
        prompt: userInput,
        callbacks: {
          onChunk: (chunk) => {
            if (chunk !== '') {
              setMessages((prev) => {
                const trimmedChunk = chunk.trim()
                const cleanedForMarkers = stripAnsiCodes(trimmedChunk)

                // Check for code block end first to prevent any further processing if found
                if (cleanedForMarkers === '<\\...>' || cleanedForMarkers.includes('<\\...>')) {
                  setCodeBlockState({
                    isActive: false,
                    type: null,
                    language: undefined,
                  })
                  return prev // Don't add the end marker as a message
                }

                // Check for code block start
                const isCodeBlockStart = cleanedForMarkers.replace(/^\s+/, '').startsWith('###')
                if (isCodeBlockStart) {
                  setCodeBlockState({
                    isActive: true,
                    type: 'diff',
                    language: undefined,
                  })
                  return [
                    ...prev,
                    {
                      type: 'assistant',
                      content: trimmedChunk,
                      isCode: true,
                      isDiff: true,
                    },
                  ]
                }

                // If we're in a code block, append to the last code block message
                if (codeBlockState.isActive) {
                  // Find the last code message by iterating from the end
                  const lastCodeMessageIndex = [...prev].reverse().findIndex((msg: Message) => msg.type === 'assistant' && msg.isCode)
                  if (lastCodeMessageIndex !== -1) {
                    const actualIndex = prev.length - 1 - lastCodeMessageIndex
                    const lastCodeMessage = prev[actualIndex]
                    return [
                      ...prev.slice(0, actualIndex),
                      {
                        ...lastCodeMessage,
                        content: `${lastCodeMessage.content}\n${chunk}`,
                        isDiff: codeBlockState.type === 'diff',
                      },
                      ...prev.slice(actualIndex + 1),
                    ]
                  }
                  // If no existing code block message found, create new one
                  return [
                    ...prev,
                    {
                      type: 'assistant',
                      content: chunk,
                      isCode: true,
                      isDiff: codeBlockState.type === 'diff',
                    },
                  ]
                }

                // For non-code content after a code block, always start a new message
                const previousMessage = prev[prev.length - 1]
                if (previousMessage?.isCode) {
                  return [
                    ...prev,
                    {
                      type: 'assistant',
                      content: chunk,
                      isCode: false,
                      isDiff: false,
                    },
                  ]
                }

                // Handle regular messages
                if (previousMessage?.type === 'assistant' && !previousMessage.isCode) {
                  const shouldCreateNewMessage =
                    previousMessage.content.trim().endsWith(':') ||
                    trimmedChunk.startsWith('verb:') ||
                    trimmedChunk.startsWith('request:') ||
                    /^\d+\./.test(trimmedChunk) ||
                    (previousMessage.content.trim().startsWith('-') && !trimmedChunk.startsWith('-'))

                  if (!shouldCreateNewMessage) {
                    return [...prev.slice(0, -1), { ...previousMessage, content: `${previousMessage.content}\n${chunk}` }]
                  }
                }

                // Create a new regular message
                return [
                  ...prev,
                  {
                    type: 'assistant',
                    content: chunk,
                    isCode: false,
                    isDiff: false,
                  },
                ]
              })
            }
          },
          onComplete: () => {
            setIsProcessing(false)
          },
          onError: (error) => {
            console.error('Error streaming from Goose:', error)
            setIsProcessing(false)

            let errorMessage = 'Sorry, I encountered an error processing your request.'
            if (error instanceof Error) {
              if (error.message.includes('command not found') || error.message.includes('ENOENT')) {
                errorMessage = 'Unable to find the FTL command. Please ensure FTL is properly installed and available in your system PATH.'
              } else if (error.message.includes('exit status')) {
                errorMessage = 'The Goose AI assistant encountered an error. Please try again with a different query.'
              }
            }

            setMessages((prev) => [...prev, { type: 'assistant', content: errorMessage }])
          },
        },
      })
    } catch (error) {
      console.error('Caught error in handleSubmit:', error)
      setIsProcessing(false)
      setMessages((prev) => [
        ...prev,
        {
          type: 'assistant',
          content: 'Sorry, an unexpected error occurred. Please try again.',
        },
      ])
    }
  }

  // Resize handlers
  const handleResizeStart = (e: React.MouseEvent) => {
    e.preventDefault()
    isDraggingRef.current = true
    document.addEventListener('mousemove', handleResizeMove)
    document.addEventListener('mouseup', handleResizeEnd)
  }

  const handleResizeMove = (e: MouseEvent) => {
    if (!isDraggingRef.current) return
    const windowWidth = window.innerWidth
    const newWidth = Math.min(Math.max(20, ((windowWidth - e.clientX) / windowWidth) * 100), 80)
    setWidth(newWidth)
  }

  const handleResizeEnd = () => {
    isDraggingRef.current = false
    document.removeEventListener('mousemove', handleResizeMove)
    document.removeEventListener('mouseup', handleResizeEnd)
  }

  // Cleanup event listeners on unmount
  useEffect(() => {
    return () => {
      document.removeEventListener('mousemove', handleResizeMove)
      document.removeEventListener('mouseup', handleResizeEnd)
    }
  }, [])

  if (!isOpen) return null

  return (
    <div
      ref={panelRef}
      style={{ width: `${width}%` }}
      className='fixed right-0 top-16 bottom-0 bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 shadow-xl flex flex-col z-50 overflow-hidden transition-width duration-75'
    >
      <div className='absolute left-0 top-0 bottom-0 w-1 cursor-ew-resize hover:bg-indigo-500 hover:opacity-50' onMouseDown={handleResizeStart}>
        <div className='absolute left-0 top-1/2 -translate-y-1/2 -translate-x-1/2 w-5 h-8 bg-indigo-500 rounded flex items-center justify-center'>
          <GripVertical className='size-3 text-white' />
        </div>
      </div>
      <div className='flex justify-between items-center p-4 border-b border-gray-200 dark:border-gray-700'>
        <h2 className='text-lg font-semibold flex items-center gap-2'>
          <GooseIcon size={20} />
          <span>Goose</span>
        </h2>
        <button type='button' onClick={onClose} className='p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700'>
          <XIcon className='size-5' />
        </button>
      </div>

      <div className='flex-1 overflow-auto p-4 space-y-4'>
        {messages.length === 0 && (
          <div className='text-gray-500 dark:text-gray-400 text-center p-4'>
            <div className='flex justify-center mb-3'>
              <GooseIcon size={32} className='text-indigo-500' />
            </div>
            <p className='mb-2'>Hi! I'm Goose, your FTL assistant.</p>
            <p>How can I help you today?</p>
          </div>
        )}
        {messages.map((message, index) => (
          <div key={index} className={classNames(message.type === 'user' ? 'flex justify-end' : 'flex justify-start')}>
            {message.type === 'assistant' ? (
              <div
                className={classNames(
                  'p-3 rounded-lg max-w-[85%]',
                  message.isDiff
                    ? 'bg-gray-900 font-mono text-xs leading-relaxed'
                    : message.isCode
                      ? 'bg-gray-800 font-mono text-xs leading-relaxed'
                      : 'bg-gray-700 font-sans',
                  'text-white',
                )}
              >
                {message.isCode || message.isDiff ? (
                  <CodeBlock content={message.content} isDiff={message.isDiff || false} language={message.language} />
                ) : (
                  <MarkdownContent content={message.content} />
                )}
              </div>
            ) : (
              <div className='bg-indigo-500 text-white p-3 rounded-lg max-w-[85%]'>
                <p>{message.content}</p>
              </div>
            )}
          </div>
        ))}
        {isProcessing && (
          <div className='p-3 rounded-lg max-w-[80%] bg-gray-100 dark:bg-gray-700'>
            <span className='animate-pulse text-lg'>...</span>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <form onSubmit={handleSubmit} className='p-4 border-t border-gray-200 dark:border-gray-700'>
        <div className='flex space-x-2'>
          <input
            type='text'
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder='Ask anything about your FTL project...'
            className='flex-1 px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-indigo-500'
          />
          <button
            type='submit'
            disabled={isProcessing}
            className={classNames(
              'px-4 py-2 bg-indigo-600 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500',
              isProcessing ? 'opacity-50 cursor-not-allowed' : 'hover:bg-indigo-500',
            )}
          >
            <SendIcon className='size-5' />
          </button>
        </div>
      </form>
    </div>
  )
}
