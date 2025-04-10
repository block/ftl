import { GripVertical, Send as SendIcon, X as XIcon } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { classNames } from '../../shared/utils'
import { GooseIcon } from './GooseIcon'
import { useGoose } from './use-goose'

type Message = {
  type: 'user' | 'assistant'
  content: string
}

export const AIAgent = ({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) => {
  const [input, setInput] = useState('')
  const [messages, setMessages] = useState<Message[]>([])
  const [isProcessing, setIsProcessing] = useState(false)
  const [width, setWidth] = useState(33) // Width as percentage
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
                const lastMessage = prev[prev.length - 1]
                const trimmedChunk = chunk.trim()

                // If chunk starts with "-", always append to previous message
                if (trimmedChunk.startsWith('-') && lastMessage?.type === 'assistant') {
                  return [...prev.slice(0, -1), { ...lastMessage, content: `${lastMessage.content}\n${chunk}` }]
                }

                // Otherwise, create new message if:
                // 1. No previous assistant message exists
                // 2. Last message ended with ":"
                // 3. Current chunk starts with a number followed by "."
                // 4. Last message was a bullet point but current isn't
                const lastMessageLines = lastMessage?.content.split('\n') || []
                const lastLine = lastMessageLines[lastMessageLines.length - 1].trim()
                const shouldCreateNewMessage =
                  !lastMessage ||
                  lastMessage.type !== 'assistant' ||
                  lastMessage.content.trim().endsWith(':') ||
                  /^\d+\./.test(trimmedChunk) ||
                  (lastLine.startsWith('-') && !trimmedChunk.startsWith('-'))

                if (lastMessage?.type === 'assistant' && !shouldCreateNewMessage) {
                  return [...prev.slice(0, -1), { ...lastMessage, content: `${lastMessage.content}\n${chunk}` }]
                }

                return [...prev, { type: 'assistant', content: chunk }]
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
              <div className='bg-gray-700 text-white p-3 rounded-lg max-w-[85%]'>
                <pre className='font-sans whitespace-pre-wrap break-words'>{message.content}</pre>
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
