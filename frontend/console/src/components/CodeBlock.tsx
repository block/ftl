import { json } from '@codemirror/lang-json'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { atomone } from '@uiw/codemirror-theme-atomone'
import { githubLight } from '@uiw/codemirror-theme-github'
import { useEffect, useRef } from 'react'
import { useUserPreferences } from '../providers/user-preferences-provider'

interface Props {
  code: string
  maxHeight?: number
}

export const CodeBlock = ({ code, maxHeight = 300 }: Props) => {
  const editorContainerRef = useRef<HTMLDivElement>(null)
  const editorViewRef = useRef<EditorView | null>(null)
  const { isDarkMode } = useUserPreferences()

  useEffect(() => {
    if (editorContainerRef.current) {
      const state = EditorState.create({
        doc: code,
        extensions: [
          EditorState.readOnly.of(true),
          isDarkMode ? atomone : githubLight,
          json(),
          EditorView.theme({
            '&': {
              minHeight: '2em',
              backgroundColor: isDarkMode ? '#282c34' : '#f6f8fa',
            },
            '.cm-scroller': {
              overflow: 'auto',
              padding: '6px',
              maxHeight: maxHeight ? `${maxHeight}px` : 'none',
            },
            '.cm-content': {
              minHeight: '2em',
            },
          }),
        ],
      })

      const view = new EditorView({
        state,
        parent: editorContainerRef.current,
      })

      editorViewRef.current = view

      return () => {
        view.destroy()
      }
    }
  }, [code, isDarkMode, maxHeight])

  return <div ref={editorContainerRef} className='rounded-md overflow-hidden border border-gray-200 dark:border-gray-600/20' />
}
