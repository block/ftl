import { useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useStatus } from '../../api/status/use-status'
import { Tabs } from '../../components/Tabs'
import { useStreamEngineEvents } from '../engine/use-stream-engine-events'
import { BuildEngineEvents } from './BuildEngineEvents'
import { ControllersList } from './ControllersList'
import { DeploymentsList } from './DeploymentsList'
import { RoutesList } from './RoutesList'
import { RunnersList } from './RunnersList'

interface Tab {
  name: string
  id: string
  count?: number
}

export const InfrastructurePage = () => {
  const status = useStatus()
  const navigate = useNavigate()
  const location = useLocation()
  const { data } = useStreamEngineEvents()
  const events = useMemo(() => (data?.pages ?? []).flatMap((page) => (Array.isArray(page) ? page : [])), [data?.pages])

  const [tabs, setTabs] = useState<Tab[]>([
    { name: 'Controllers', id: 'controllers' },
    { name: 'Runners', id: 'runners' },
    { name: 'Deployments', id: 'deployments' },
    { name: 'Routes', id: 'routes' },
    { name: 'Build Engine Events', id: 'build-engine-events' },
  ])

  const currentTab = location.pathname.split('/').pop()

  useEffect(() => {
    if (!status.data) {
      return
    }

    setTabs((prevTabs: Tab[]) =>
      prevTabs.map((tab: Tab) => {
        switch (tab.id) {
          case 'controllers':
            return { ...tab, count: status.data.controllers.length }
          case 'runners':
            return { ...tab, count: status.data.runners.length }
          case 'deployments':
            return { ...tab, count: status.data.deployments.length }
          case 'routes':
            return { ...tab, count: status.data.routes.length }
          case 'build-engine-events': {
            const { count: _, ...rest } = tab
            return rest
          }
          default:
            return tab
        }
      }),
    )
  }, [status.data])

  const renderTabContent = () => {
    switch (currentTab) {
      case 'controllers':
        return <ControllersList controllers={status.data?.controllers || []} />
      case 'runners':
        return <RunnersList runners={status.data?.runners || []} />
      case 'deployments':
        return <DeploymentsList deployments={status.data?.deployments || []} />
      case 'routes':
        return <RoutesList routes={status.data?.routes || []} />
      case 'build-engine-events':
        return <BuildEngineEvents events={events} />
      default:
        return <></>
    }
  }

  const handleTabClick = (tabId: string) => {
    navigate(`/infrastructure/${tabId}`)
  }

  return (
    <div className='h-full flex flex-col px-6'>
      <div className='flex-none'>
        <Tabs tabs={tabs} initialTabId={currentTab} onTabClick={handleTabClick} />
      </div>
      <div className='flex-1 overflow-auto mt-2'>{renderTabContent()}</div>
    </div>
  )
}
