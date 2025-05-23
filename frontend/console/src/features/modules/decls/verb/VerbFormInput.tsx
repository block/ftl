import { Copy01Icon } from 'hugeicons-react'
import { useEffect, useRef } from 'react'
import { Button } from '../../../../shared/components/Button'
import { Loader } from '../../../../shared/components/Loader'

export const VerbFormInput = ({
  requestType,
  path,
  setPath,
  requestPath,
  readOnly,
  onSubmit,
  handleCopyButton,
  isLoading = false,
}: {
  requestType: string
  path: string
  setPath: (path: string) => void
  requestPath: string
  readOnly: boolean
  onSubmit: (path: string) => void
  handleCopyButton?: () => void
  isLoading?: boolean
}) => {
  const formRef = useRef<HTMLFormElement>(null)

  const handleSubmit: React.FormEventHandler<HTMLFormElement> = async (event) => {
    event.preventDefault()
    onSubmit(path)
  }

  const shortcutText = `Send ${window.navigator.userAgent.includes('Mac') ? '⌘ + ⏎' : 'Ctrl + ⏎'}`

  useEffect(() => {
    const handleKeydown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
        event.preventDefault()
        formRef.current?.dispatchEvent(new Event('submit', { cancelable: true, bubbles: true }))
      }
    }

    window.addEventListener('keydown', handleKeydown)

    return () => {
      window.removeEventListener('keydown', handleKeydown)
    }
  }, [path, readOnly, onSubmit])

  return (
    <form ref={formRef} onSubmit={handleSubmit} className='rounded-lg'>
      <div className='flex rounded-md'>
        <span id='call-type' className='inline-flex items-center rounded-l-md border border-r-0 border-gray-300 dark:border-gray-500 px-3 ml-4 sm:text-sm'>
          {requestType}
        </span>
        <input
          type='text'
          name='request-path'
          id='request-path'
          className='block w-full min-w-0 flex-1 rounded-none rounded-r-md border-0 py-1.5 dark:bg-transparent ring-1 ring-inset ring-gray-300 dark:ring-gray-500 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6'
          value={path}
          readOnly={readOnly}
          onChange={(event) => setPath(event.target.value)}
        />
        <Button variant='primary' size='md' type='submit' title={shortcutText} className='mx-2' disabled={isLoading}>
          {isLoading ? <Loader className='size-5' /> : 'Send'}
        </Button>
        <Button variant='secondary' size='md' type='button' title='Copy' onClick={handleCopyButton} className='mr-2'>
          <Copy01Icon className='size-5' />
        </Button>
      </div>
      {!readOnly && <span className='ml-4 text-xs text-gray-500'>{requestPath}</span>}
    </form>
  )
}
