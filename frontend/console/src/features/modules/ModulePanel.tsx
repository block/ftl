import { CellsIcon, CodeIcon } from 'hugeicons-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import { ResizablePanels } from '../../shared/components/ResizablePanels'
import { headerForNode } from '../graph/RightPanelHeader'
import { type FTLNode, panelsForNode } from '../graph/graph-utils'
import { ModuleGraph } from './ModuleGraph'
import { useStreamModules } from './hooks/use-stream-modules'
import { Schema } from './schema/Schema'

export const ModulePanel = () => {
  const { moduleName } = useParams()
  const modules = useStreamModules()
  const ref = useRef<HTMLDivElement>(null)
  const [selectedNode, setSelectedNode] = useState<FTLNode | null>(null)
  const [selectedModuleName, setSelectedModuleName] = useState<string | null>(null)

  const module = useMemo(() => {
    if (!modules?.data) {
      return
    }
    return modules.data.modules.find((module) => module.name === moduleName)
  }, [modules?.data, moduleName])

  useEffect(() => {
    ref?.current?.parentElement?.scrollTo({ top: 0, behavior: 'smooth' })
  }, [moduleName])

  const handleNodeTapped = (node: FTLNode | null, nodeName: string | null) => {
    setSelectedNode(node)
    setSelectedModuleName(nodeName)
  }

  if (!module) return

  const mainContent = (
    <div ref={ref} className='p-4 h-[calc(100vh-64px)] flex flex-col'>
      <div className='flex-1 min-h-0 grid grid-rows-2 gap-4'>
        <div className='border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden'>
          <div className='flex items-center gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800'>
            <CodeIcon className='size-4 text-gray-500 dark:text-gray-400' />
            <span className='text-sm font-medium text-gray-700 dark:text-gray-200'>Schema</span>
          </div>
          <div className='p-4 overflow-auto h-[calc(100%-40px)]'>
            <Schema schema={module.schema} moduleName={module.name} />
          </div>
        </div>
        <div className='border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden'>
          <div className='flex items-center gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800'>
            <CellsIcon className='size-4 text-gray-500 dark:text-gray-400' />
            <span className='text-sm font-medium text-gray-700 dark:text-gray-200'>Graph</span>
          </div>
          <div className='h-[calc(100%-40px)]'>
            <ModuleGraph module={module} onTapped={handleNodeTapped} />
          </div>
        </div>
      </div>
    </div>
  )

  return (
    <div className='h-[calc(100vh-64px)] flex'>
      <div className='flex-1 min-h-0'>
        <ResizablePanels
          mainContent={mainContent}
          rightPanelHeader={headerForNode(selectedNode, selectedModuleName)}
          rightPanelPanels={panelsForNode(selectedNode, selectedModuleName)}
        />
      </div>
    </div>
  )
}
