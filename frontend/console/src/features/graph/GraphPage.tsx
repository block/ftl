import { useState } from 'react'
import { Loader } from '../../shared/components/Loader'
import { ResizablePanels } from '../../shared/components/ResizablePanels'
import { useModules } from '../modules/hooks/use-modules'
import { Timeline } from '../timeline/Timeline'
import { GraphPane } from './GraphPane'
import { headerForNode } from './RightPanelHeader'
import { type FTLNode, panelsForNode } from './graph-utils'

export const GraphPage = () => {
  const modules = useModules()
  const [selectedNode, setSelectedNode] = useState<FTLNode | null>(null)
  const [selectedModuleName, setSelectedModuleName] = useState<string | null>(null)

  if (!modules.isSuccess) {
    return (
      <div className='flex justify-center items-center h-full'>
        <Loader />
      </div>
    )
  }

  const handleNodeTapped = (node: FTLNode | null, moduleName: string | null) => {
    setSelectedNode(node)
    setSelectedModuleName(moduleName)
  }

  return (
    <div className='flex h-full'>
      <ResizablePanels
        mainContent={<GraphPane onTapped={handleNodeTapped} />}
        rightPanelHeader={headerForNode(selectedNode, selectedModuleName)}
        rightPanelPanels={panelsForNode(selectedNode, selectedModuleName)}
        bottomPanelContent={<Timeline timeSettings={{ isTailing: true, isPaused: false }} filters={[]} />}
      />
    </div>
  )
}
