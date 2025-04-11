import { Bot as BotIcon } from 'lucide-react'
import type React from 'react'
import { useEffect } from 'react'
import { classNames } from '../../utils'

type AIInputProps = {
  isOpen: boolean
  onToggle: () => void
}

export const AIInput: React.FC<AIInputProps> = ({ isOpen, onToggle }) => {
  const shortcutText = window.navigator.userAgent.includes('Mac') ? '⌘ + I' : 'Ctrl + I'

  useEffect(() => {
    const handleKeydown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'i') {
        const tagName = (event.target as HTMLElement)?.tagName?.toLowerCase()
        if (tagName === 'input' || tagName === 'textarea') {
          return
        }

        event.preventDefault()
        onToggle()
      }
    }

    window.addEventListener('keydown', handleKeydown)

    return () => {
      window.removeEventListener('keydown', handleKeydown)
    }
  }, [onToggle])

  return (
    <button
      type='button'
      onClick={onToggle}
      className={classNames(
        isOpen ? 'bg-indigo-700' : 'hover:bg-indigo-500 hover:bg-opacity-75',
        'text-white rounded-md px-3 py-2 pr-16 ml-4 text-sm font-medium relative',
      )}
      aria-label='Toggle Goose assistant'
    >
      <div className='flex items-center'>
        <BotIcon className='size-5' />
        <span className='sr-only'>Goose AI Assistant</span>
      </div>
      <div className='absolute inset-y-0 right-2 flex items-center'>
        <span className='text-indigo-200 text-xs bg-indigo-600 px-2 py-1 rounded-md'>{shortcutText}</span>
      </div>
    </button>
  )
}
