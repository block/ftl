import { useContext, useEffect, useState } from 'react'
import { ConsoleService } from '../../../../protos/xyz/block/ftl/console/v1/console_connect'
import type { Config } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import { Button } from '../../../../shared/components/Button'
import { Checkbox } from '../../../../shared/components/Checkbox'
import { CodeEditor } from '../../../../shared/components/CodeEditor'
import { ResizablePanels } from '../../../../shared/components/ResizablePanels'
import { useClient } from '../../../../shared/hooks/use-client'
import { NotificationType, NotificationsContext } from '../../../../shared/providers/notifications-provider'
import { declIcon, graphUrlForRef } from '../../module.utils'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { configPanels } from './ConfigRightPanels'

export const ConfigPanel = ({ config, moduleName, declName }: { config: Config; moduleName: string; declName: string }) => {
  const client = useClient(ConsoleService)
  const [configValue, setConfigValue] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isJsonMode, setIsJsonMode] = useState(false)
  const notification = useContext(NotificationsContext)

  useEffect(() => {
    handleGetConfig()
  }, [moduleName, declName])

  const handleGetConfig = () => {
    setIsLoading(true)
    client.getConfig({ module: moduleName, name: declName }).then((resp) => {
      setConfigValue(new TextDecoder().decode(resp.value))
      setIsLoading(false)
    })
  }

  const handleSetConfig = () => {
    setIsLoading(true)
    let valueToSend = configValue

    if (isJsonMode) {
      try {
        JSON.parse(configValue)
      } catch (e) {
        notification?.showNotification({
          title: 'Invalid JSON',
          message: 'Please enter valid JSON',
          type: NotificationType.Error,
        })
        setIsLoading(false)
        return
      }
    } else {
      valueToSend = configValue
    }

    client
      .setConfig({
        module: moduleName,
        name: declName,
        value: new TextEncoder().encode(valueToSend),
      })
      .then(() => {
        setIsLoading(false)
        notification?.showNotification({
          title: 'Config updated',
          message: 'Config updated successfully',
          type: NotificationType.Success,
        })
      })
      .catch((error) => {
        setIsLoading(false)
        notification?.showNotification({
          title: 'Failed to update config',
          message: error.message,
          type: NotificationType.Error,
        })
      })
  }

  if (!config) {
    return null
  }
  const decl = config.config
  if (!decl) {
    return null
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <div className=''>
              <PanelHeader title='Config' declRef={`${moduleName}.${declName}`} exported={false} comments={decl.comments} />
              <CodeEditor value={configValue} onTextChanged={setConfigValue} />
              <div className='mt-2 flex items-center justify-between'>
                <Checkbox checked={isJsonMode} onChange={(e) => setIsJsonMode(e.target.checked)} label='JSON mode' />
                <div className='space-x-2 flex flex-nowrap'>
                  <Button onClick={handleSetConfig} disabled={isLoading}>
                    Save
                  </Button>
                  <Button onClick={handleGetConfig} disabled={isLoading}>
                    Refresh
                  </Button>
                </div>
              </div>
            </div>
          </div>
        }
        rightPanelHeader={
          <RightPanelHeader Icon={declIcon('config', decl)} title={declName} url={graphUrlForRef(`${moduleName}.${declName}`)} urlHoverText='View in graph' />
        }
        rightPanelPanels={configPanels(moduleName, config)}
        storageKeyPrefix='configPanel'
      />
    </div>
  )
}
