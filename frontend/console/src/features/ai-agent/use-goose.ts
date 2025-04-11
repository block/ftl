import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { useMutation } from '@tanstack/react-query'

import { ConsoleService } from '../../protos/xyz/block/ftl/console/v1/console_connect'
import { ExecuteGooseRequest, ExecuteGooseResponse_Source } from '../../protos/xyz/block/ftl/console/v1/console_pb'

const transport = createConnectTransport({
  baseUrl: window.location.origin,
})

export type GooseStreamCallbacks = {
  onChunk?: (chunk: string) => void
  onError?: (error: Error) => void
  onComplete?: () => void
}

export type ExecuteGooseParams = {
  prompt: string
  callbacks?: GooseStreamCallbacks
}

export const useGoose = () => {
  return useMutation({
    mutationFn: async ({ prompt, callbacks }: ExecuteGooseParams) => {
      const client = createClient(ConsoleService, transport)
      const request = new ExecuteGooseRequest({ prompt })

      try {
        let completeResponse = ''
        const responseStream = client.executeGoose(request)

        try {
          for await (const chunk of responseStream) {
            if (chunk.source === ExecuteGooseResponse_Source.COMPLETION) {
              callbacks?.onComplete?.()
              break
            }

            if (chunk.source === ExecuteGooseResponse_Source.STDOUT) {
              const chunkText = chunk.response
              if (chunkText && chunkText.trim() !== '') {
                completeResponse += chunkText
                callbacks?.onChunk?.(chunkText)
              }
            }
          }

          // Stream completed normally
          callbacks?.onComplete?.()
        } catch (streamError) {
          console.error('Streaming error:', streamError)
          if (callbacks?.onError && streamError instanceof Error) {
            callbacks.onError(streamError)
          }
          callbacks?.onComplete?.() // Ensure onComplete is called even on error
          throw streamError
        }

        return completeResponse
      } catch (error) {
        console.error('Error with Goose API:', error)
        if (callbacks?.onError && error instanceof Error) {
          callbacks.onError(error)
        }
        callbacks?.onComplete?.() // Ensure onComplete is called in case of API error
        throw error
      }
    },
  })
}
