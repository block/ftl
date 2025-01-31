import { useState } from 'react'
import { Loader } from '../../shared/components/Loader'
import { ResizablePanels } from '../../shared/components/ResizablePanels'
import { useStreamModules } from '../modules/hooks/use-stream-modules'
import { Timeline } from '../timeline/Timeline'
import { GraphPane } from './GraphPane'
import { headerForNode } from './RightPanelHeader'
import { type FTLNode, panelsForNode } from './graph-utils'

export const GraphPage = () => {
  const { data, isLoading } = useStreamModules()
  const [selectedNode, setSelectedNode] = useState<FTLNode | null>(null)
  const [selectedModuleName, setSelectedModuleName] = useState<string | null>(null)

  if (isLoading) {
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
        mainContent={<GraphPane modules={data} onTapped={handleNodeTapped} />}
        rightPanelHeader={headerForNode(selectedNode, selectedModuleName)}
        rightPanelPanels={panelsForNode(selectedNode, selectedModuleName)}
        bottomPanelContent={<Timeline timeSettings={{ isTailing: true, isPaused: false }} filters={[]} />}
      />
    </div>
  )
}
