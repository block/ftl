import { Controls, type Edge, ReactFlow as Flow, type Node, ReactFlowProvider } from '@xyflow/react'
import dagre from 'dagre'
import { useCallback, useMemo } from 'react'
import type React from 'react'
import type { Edges, Module } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { useUserPreferences } from '../../../shared/providers/user-preferences-provider'
import { DeclNode } from '../../graph/DeclNode'
import { getNodeBackgroundColor } from '../../graph/graph-styles'
import { useModules } from '../hooks/use-modules'
import { getVerbType } from '../module.utils'
import { declSchemaFromModules } from '../schema/schema.utils'
import '@xyflow/react/dist/style.css'
import '../../graph/graph.css'

const NODE_TYPES = {
  declNode: DeclNode,
}

interface DeclGraphPaneProps {
  edges?: Edges
  declName: string
  declType: string
  moduleName: string
  height?: string
}

/**
 * Helper function to find a verb in modules and get its type
 */
const getVerbNodeType = (modules: Module[], moduleName: string, verbName: string, defaultType: string): string => {
  if (defaultType !== 'verb') {
    return defaultType
  }

  const module = modules.find((m) => m.name === moduleName)
  if (!module) {
    return defaultType
  }

  const verb = module.verbs.find((v) => v.verb?.name === verbName)
  if (!verb) {
    return defaultType
  }

  return getVerbType(verb)
}

// Dagre layout function for decl graph
const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'LR') => {
  const dagreGraph = new dagre.graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))

  const nodeWidth = 160
  const nodeHeight = 20

  dagreGraph.setGraph({
    rankdir: direction,
    nodesep: 50,
    ranksep: 100,
  })

  // Add nodes to dagre
  for (const node of nodes) {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight })
  }

  // Add edges to dagre
  for (const edge of edges) {
    dagreGraph.setEdge(edge.source, edge.target)
  }

  // Apply layout
  dagre.layout(dagreGraph)

  // Get positions from dagre
  for (const node of nodes) {
    const nodeWithPosition = dagreGraph.node(node.id)
    node.position = {
      x: nodeWithPosition.x - nodeWidth / 2,
      y: nodeWithPosition.y - nodeHeight / 2,
    }
  }

  return { nodes, edges }
}

const createNode = (id: string, label: string, nodeType: string, isDarkMode: boolean, isCenter: boolean): Node => ({
  id,
  type: 'declNode',
  position: { x: 0, y: 0 },
  data: {
    title: label,
    selected: isCenter,
    type: 'declNode',
    nodeType,
    item: { name: label },
    style: {
      backgroundColor: getNodeBackgroundColor(isDarkMode, nodeType),
    },
  },
})

const createEdge = (source: string, target: string, isDarkMode: boolean): Edge => ({
  id: `edge-${source}->${target}`,
  source,
  target,
  type: 'default',
  animated: true,
  style: {
    stroke: isDarkMode ? '#EC4899' : '#F472B6',
    strokeWidth: 2,
  },
})

export const DeclGraphPane: React.FC<DeclGraphPaneProps> = ({ edges, declName, declType, moduleName, height }) => {
  if (!edges) return null

  const { isDarkMode } = useUserPreferences()
  const { data: modulesData } = useModules()

  const { nodes, graphEdges } = useMemo(() => {
    const nodes: Node[] = []
    const graphEdges: Edge[] = []
    const modules = modulesData?.modules || []

    // Create center node (the current decl)
    const centerId = `${moduleName}.${declName}`

    // Get the correct node type for verbs
    const centerNodeType = getVerbNodeType(modules, moduleName, declName, declType)
    nodes.push(createNode(centerId, declName, centerNodeType, isDarkMode, true))

    // Create nodes and edges for inbound references
    for (const ref of edges.in) {
      const id = `${ref.module}.${ref.name}`
      const schema = declSchemaFromModules(ref.module, ref.name, modules)
      const nodeType = getVerbNodeType(modules, ref.module, ref.name, schema?.declType || declType)

      nodes.push(createNode(id, ref.name, nodeType, isDarkMode, false))
      graphEdges.push(createEdge(id, centerId, isDarkMode))
    }

    // Create nodes and edges for outbound references
    for (const ref of edges.out) {
      const id = `${ref.module}.${ref.name}`
      const schema = declSchemaFromModules(ref.module, ref.name, modules)
      const nodeType = getVerbNodeType(modules, ref.module, ref.name, schema?.declType || declType)

      nodes.push(createNode(id, ref.name, nodeType, isDarkMode, false))
      graphEdges.push(createEdge(centerId, id, isDarkMode))
    }

    const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(nodes, graphEdges)
    return { nodes: layoutedNodes, graphEdges: layoutedEdges }
  }, [edges, declName, declType, moduleName, isDarkMode, modulesData?.modules])

  // Calculate height based on number of nodes
  const calculatedHeight = useMemo(() => {
    if (height) return height
    const nodeCount = Math.max(edges.in.length, edges.out.length, 1)
    // 20px node height + 50px node separation + 40px padding
    return `${nodeCount * (20 + 50) + 60}px`
  }, [edges.in.length, edges.out.length, height])

  const onInit = useCallback(() => {
    // ReactFlow instance initialization if needed
  }, [])

  return (
    <ReactFlowProvider key={`${moduleName}.${declName}`}>
      <div className={isDarkMode ? 'dark' : 'light'} style={{ width: '100%', height: calculatedHeight }}>
        <Flow
          nodes={nodes}
          edges={graphEdges}
          nodeTypes={NODE_TYPES}
          onInit={onInit}
          fitView
          minZoom={0.1}
          maxZoom={2}
          proOptions={{ hideAttribution: true }}
          nodesDraggable={false}
          nodesConnectable={false}
          colorMode={isDarkMode ? 'dark' : 'light'}
        >
          <Controls showInteractive={false} showZoom={false} />
        </Flow>
      </div>
    </ReactFlowProvider>
  )
}
