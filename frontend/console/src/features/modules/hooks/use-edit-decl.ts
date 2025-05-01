import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { useMutation } from '@tanstack/react-query'
import { useContext } from 'react'

import { ConsoleService } from '../../../protos/xyz/block/ftl/console/v1/console_connect'
import { OpenFileInEditorRequest } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { NotificationType, NotificationsContext } from '../../../shared/providers/notifications-provider'

// Reusable transport, similar to use-goose.ts
const transport = createConnectTransport({
  baseUrl: window.location.origin,
})

export type EditDeclParams = {
  editor: string
  path: string
  line: number
  column: number
}

export const useEditDecl = () => {
  const { showNotification } = useContext(NotificationsContext)

  return useMutation<void, Error, EditDeclParams>({
    mutationFn: async ({ editor, path, line, column }: EditDeclParams): Promise<void> => {
      const client = createClient(ConsoleService, transport)
      const request = new OpenFileInEditorRequest({
        editor,
        path,
        line,
        column,
      })

      await client.openFileInEditor(request)
    },
    onError: (error) => {
      console.error('Error calling OpenFileInEditor:', error)
      showNotification({
        title: 'Failed to Open File',
        message: error.message || 'An unknown error occurred',
        type: NotificationType.Error,
      })
    },
  })
}
