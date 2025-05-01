import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/react'
import { ArrowDown01Icon } from 'hugeicons-react'
import { Fragment, useEffect, useMemo } from 'react'
import { useLocalStorage } from '../../../shared/hooks/use-local-storage'
import { useInfo } from '../../../shared/providers/info-provider'
import { classNames } from '../../../shared/utils'
import { type EditDeclParams, useEditDecl } from '../hooks/use-edit-decl'

const editorDisplayNames: Record<string, string> = {
  vscode: 'VSCode',
  intellij: 'IntelliJ',
  cursor: 'Cursor',
  zed: 'Zed',
  github: 'GitHub',
}

const allEditorKeys = Object.keys(editorDisplayNames)

export interface EditDeclButtonProps {
  path: string
  line: number
  column: number
}

export const EditDeclButton: React.FC<EditDeclButtonProps> = ({ path, line, column }) => {
  const { mutate: editDecl, isPending } = useEditDecl()
  const [lastEditor, setLastEditor] = useLocalStorage<string | null>('ftl-console-last-editor', null)
  const info = useInfo()

  const availableEditorKeys = useMemo(() => (info.isLocalDev ? allEditorKeys : ['github']), [info.isLocalDev])

  useEffect(() => {
    if (!lastEditor || !availableEditorKeys.includes(lastEditor)) {
      setLastEditor(availableEditorKeys[0])
    }
  }, [availableEditorKeys, lastEditor, setLastEditor])

  const effectiveEditor = lastEditor && availableEditorKeys.includes(lastEditor) ? lastEditor : availableEditorKeys[0]

  const handleEditorSelect = (editor: string) => {
    const params: EditDeclParams = { editor, path, line, column }
    editDecl(params, {
      onSuccess: () => {
        setLastEditor(editor)
      },
    })
  }

  const handlePrimaryAction = () => {
    handleEditorSelect(effectiveEditor)
  }

  const menuItemsContent = (
    <MenuItems
      transition
      anchor='bottom end'
      className={classNames(
        'absolute right-0 z-10 mt-2 w-40 origin-top-right rounded-md shadow-lg py-1',
        'bg-white dark:bg-gray-900 ring-1 ring-black dark:ring-white/10 ring-opacity-5',
        'transition focus:outline-none',
        'data-[closed]:scale-95 data-[closed]:transform data-[closed]:opacity-0',
        'data-[enter]:duration-100 data-[enter]:ease-out',
        'data-[leave]:duration-75 data-[leave]:ease-in',
      )}
    >
      {availableEditorKeys.map((editor) => (
        <MenuItem key={editor} as={Fragment}>
          {({ active }) => (
            <button
              type='button'
              onClick={() => handleEditorSelect(editor)}
              className={classNames(
                active ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white' : 'text-gray-700 dark:text-gray-200',
                editor === lastEditor ? 'font-semibold' : '',
                'block w-full px-4 py-2 text-left text-sm',
              )}
            >
              {editorDisplayNames[editor] ?? editor}
            </button>
          )}
        </MenuItem>
      ))}
    </MenuItems>
  )

  if (!info.isLocalDev) {
    return (
      <button
        type='button'
        onClick={() => handleEditorSelect('github')}
        disabled={isPending}
        title={'Edit with GitHub'}
        className={classNames(
          'relative inline-flex items-center gap-x-1.5 rounded-md px-3 py-1',
          'bg-white dark:bg-gray-800 ring-1 ring-gray-300 dark:ring-gray-700 ring-inset',
          'text-sm font-semibold text-gray-900 dark:text-white',
          'hover:bg-gray-100 dark:hover:bg-gray-700',
          'focus:z-10 focus:outline-none focus:ring-2 focus:ring-indigo-500',
          'disabled:opacity-50 disabled:cursor-not-allowed',
        )}
      >
        {editorDisplayNames.github}
      </button>
    )
  }

  if (!lastEditor || !allEditorKeys.includes(lastEditor)) {
    return (
      <Menu as='div' className='relative inline-block text-left'>
        <div>
          <MenuButton
            disabled={isPending}
            className={classNames(
              'inline-flex items-center justify-center w-full rounded-md px-3 py-1 text-sm font-semibold shadow-sm',
              'bg-white dark:bg-gray-800 ring-1 ring-gray-300 dark:ring-gray-700 ring-inset',
              'text-gray-900 dark:text-white',
              'hover:bg-gray-50 dark:hover:bg-gray-700',
              'disabled:opacity-50 disabled:cursor-not-allowed',
            )}
          >
            Open...
            <ArrowDown01Icon className='-mr-1 ml-2 h-5 w-5 text-gray-400' aria-hidden='true' />
          </MenuButton>
        </div>
        {menuItemsContent}
      </Menu>
    )
  }

  const buttonText = editorDisplayNames[lastEditor]
  return (
    <div
      className={classNames(
        'relative z-0 inline-flex rounded-md shadow-sm',
        'bg-white dark:bg-gray-800 ring-1 ring-gray-300 dark:ring-gray-700 ring-inset',
        'text-sm font-semibold text-gray-900 dark:text-white',
        'disabled:opacity-50',
      )}
      aria-disabled={isPending}
    >
      <button
        type='button'
        onClick={handlePrimaryAction}
        disabled={isPending}
        title={`Edit with ${buttonText}`}
        className={classNames(
          'relative inline-flex items-center gap-x-1.5 rounded-l-md px-3 py-1',
          'hover:bg-gray-100 dark:hover:bg-gray-700',
          'focus:z-10 focus:outline-none focus:ring-2 focus:ring-indigo-500',
          'disabled:cursor-not-allowed',
        )}
      >
        {buttonText}
      </button>
      <Menu as='div' className='relative -ml-px block'>
        <MenuButton
          disabled={isPending}
          className={classNames(
            'relative inline-flex items-center rounded-r-md px-2 py-1',
            'text-gray-500 dark:text-gray-400',
            'hover:bg-gray-100 dark:hover:bg-gray-700',
            'focus:z-10 focus:outline-none focus:ring-2 focus:ring-indigo-500',
            'disabled:cursor-not-allowed',
          )}
        >
          <span className='sr-only'>Open options</span>
          <ArrowDown01Icon className='h-5 w-5' aria-hidden='true' />
        </MenuButton>
        {menuItemsContent}
      </Menu>
    </div>
  )
}
