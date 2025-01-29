import { Background, BackgroundVariant, Controls, type Edge, ReactFlow as Flow, type Node, ReactFlowProvider } from '@xyflow/react'
import { useCallback, useMemo, useState } from 'react'
import type React from 'react'
import type { Module } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { Topology } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { useUserPreferences } from '../../shared/providers/user-preferences-provider'
import { DeclNode } from '../graph/DeclNode'
import type { FTLNode } from '../graph/graph-utils'
import { getGraphData, getModuleLayoutedElements } from '../graph/graph-utils'
import '@xyflow/react/dist/style.css'

const NODE_TYPES = {
  declNode: DeclNode,
}

interface ModuleGraphProps {
  module: Module
  onTapped?: (item: FTLNode | null, moduleName: string | null) => void
}

export const ModuleGraph = ({ module, onTapped }: ModuleGraphProps) => {
  const { isDarkMode } = useUserPreferences()
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)

  const { nodes, edges } = useMemo(() => {
    // Create a single-module data structure that matches what getGraphData expects
    const moduleData = {
      modules: [module],
      topology: new Topology({ levels: [] }),
      isSuccess: true,
    }
    const { nodes, edges } = getGraphData(moduleData, isDarkMode, {}, selectedNodeId)

    // Filter out the group node since we don't want it in the module view
    return {
      nodes: nodes.filter((node) => node.type !== 'groupNode'),
      edges,
    }
  }, [module, isDarkMode, selectedNodeId])

  const { nodes: layoutedNodes, edges: layoutedEdges } = useMemo(() => {
    if (!nodes.length) return { nodes: [], edges: [] }
    return getModuleLayoutedElements(nodes, edges)
  }, [nodes, edges])

  const onNodeClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      setSelectedNodeId(node.id)
      onTapped?.(node.data?.item as FTLNode, node.id)
    },
    [onTapped],
  )

  const onEdgeClick = useCallback(
    (_event: React.MouseEvent, edge: Edge) => {
      const sourceNode = layoutedNodes.find((n) => n.id === edge.source)
      const targetNode = layoutedNodes.find((n) => n.id === edge.target)

      if (sourceNode?.id === selectedNodeId || targetNode?.id === selectedNodeId) {
        setSelectedNodeId(null)
        onTapped?.(null, null)
      } else {
        setSelectedNodeId(sourceNode?.id || null)
        onTapped?.((sourceNode?.data?.item as FTLNode) || null, sourceNode?.id || null)
      }
    },
    [onTapped, layoutedNodes, selectedNodeId],
  )

  const onPaneClick = useCallback(() => {
    setSelectedNodeId(null)
    onTapped?.(null, null)
  }, [onTapped])

  return (
    <ReactFlowProvider key={module.name}>
      <div className={isDarkMode ? 'dark' : 'light'} style={{ width: '100%', height: '100%', position: 'relative' }}>
        <Flow
          nodes={layoutedNodes}
          edges={layoutedEdges}
          nodeTypes={NODE_TYPES}
          onNodeClick={onNodeClick}
          onEdgeClick={onEdgeClick}
          onPaneClick={onPaneClick}
          fitView
          minZoom={0.1}
          maxZoom={2}
          proOptions={{ hideAttribution: true }}
          nodesDraggable={false}
          nodesConnectable={false}
          colorMode={isDarkMode ? 'dark' : 'light'}
        >
          <Background variant={BackgroundVariant.Dots} gap={12} size={1} />
          <Controls />
        </Flow>
      </div>
    </ReactFlowProvider>
  )
}
