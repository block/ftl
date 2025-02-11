import { useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Tabs } from '../../shared/components/Tabs'
import { useStreamEngineEvents } from '../engine/hooks/use-stream-engine-events'
import { useStreamModules } from '../modules/hooks/use-stream-modules'
import { BuildEngineEvents } from './BuildEngineEvents'
import { DeploymentsList } from './DeploymentsList'

interface Tab {
  name: string
  id: string
  count?: number
}

export const InfrastructurePage = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { data: modulesData } = useStreamModules()
  const { data } = useStreamEngineEvents()
  const events = useMemo(() => (data?.pages ?? []).flatMap((page) => (Array.isArray(page) ? page : [])), [data?.pages])

  const [tabs, setTabs] = useState<Tab[]>([
    { name: 'Deployments', id: 'deployments' },
    { name: 'Build Engine Events', id: 'build-engine-events' },
  ])

  const currentTab = location.pathname.split('/').pop()

  const modules = useMemo(() => modulesData?.modules.filter((module) => module.name !== 'builtin') ?? [], [modulesData])

  useEffect(() => {
    setTabs((prevTabs: Tab[]) =>
      prevTabs.map((tab: Tab) => {
        switch (tab.id) {
          case 'deployments': {
            return { ...tab, count: modules.length }
          }
          case 'build-engine-events': {
            const { count: _, ...rest } = tab
            return rest
          }
          default:
            return tab
        }
      }),
    )
  }, [modulesData])

  const renderTabContent = () => {
    switch (currentTab) {
      case 'deployments':
        return <DeploymentsList modules={modules} />
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
