import type React from 'react'
import { useEffect } from 'react'
import { GooseIcon } from '../../../features/ai-agent/GooseIcon'
import { HoverPopup } from '../../components/HoverPopup'
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
    <HoverPopup popupContent={shortcutText} position='bottom'>
      <button
        type='button'
        onClick={onToggle}
        className={classNames(isOpen ? 'bg-indigo-700' : 'hover:bg-indigo-500 hover:bg-opacity-75', 'text-white rounded-md p-2 ml-4')}
        aria-label='Toggle Goose assistant'
      >
        <div className='flex items-center justify-center'>
          <GooseIcon size={20} fill='white' />
          <span className='sr-only'>Goose AI Assistant</span>
        </div>
      </button>
    </HoverPopup>
  )
}
