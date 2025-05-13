import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Loader } from '../../shared/components/Loader'
import { ResizablePanels } from '../../shared/components/ResizablePanels'
import { useStreamModules } from '../modules/hooks/use-stream-modules'
import { declFromModules } from '../modules/module.utils'
import { Timeline } from '../timeline/Timeline'
import { GraphPane } from './GraphPane'
import { HeaderForNode } from './RightPanelHeader'
import { type FTLNode, nodeId, panelsForNode } from './graph-utils'

const SHOW_TIMELINE = false

export const GraphPage = () => {
  const { data, isLoading } = useStreamModules()
  const [selectedNode, setSelectedNode] = useState<FTLNode | null>(null)
  const [selectedRef, setSelectedRef] = useState<string | null>(null)
  const { moduleName, declName } = useParams()
  const navigate = useNavigate()

  useEffect(() => {
    if (moduleName && declName && data?.modules) {
      const declTypes = ['verb', 'config', 'secret', 'database', 'topic'] as const
      const decl = declTypes.reduce<FTLNode | undefined>((found, type) => {
        if (found) return found
        return declFromModules(moduleName, type, declName, data.modules)
      }, undefined)

      if (decl) {
        setSelectedNode(decl)
        setSelectedRef(`${moduleName}.${declName}`)
      }
    }
  }, [moduleName, declName, data?.modules])

  if (isLoading) {
    return (
      <div className='flex justify-center items-center h-full'>
        <Loader />
      </div>
    )
  }

  const handleNodeTapped = (node: FTLNode | null, ref: string | null) => {
    setSelectedNode(node)
    setSelectedRef(ref)

    // Update URL when node is selected
    if (node && ref) {
      const [moduleName, declName] = ref.split('.')
      // Only include declName in URL if it exists
      navigate(declName ? `/graph/${moduleName}/${declName}` : `/graph/${moduleName}`)
    } else {
      navigate('/graph')
    }
  }

  // Get the selected node ID for the graph
  const selectedNodeId = selectedRef ? nodeId(selectedRef.split('.')[0], selectedRef.split('.')[1]) : null

  return (
    <div className='flex h-full'>
      <ResizablePanels
        mainContent={<GraphPane modules={data} onTapped={handleNodeTapped} selectedNodeId={selectedNodeId} />}
        rightPanelHeader={HeaderForNode(selectedNode, selectedRef)}
        rightPanelPanels={panelsForNode(selectedNode, selectedRef)}
        bottomPanelContent={SHOW_TIMELINE ? <Timeline timeSettings={{ isTailing: true, isPaused: false }} filters={[]} /> : undefined}
      />
    </div>
  )
}
