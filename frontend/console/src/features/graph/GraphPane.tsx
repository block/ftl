import type { ElementDefinition } from 'cytoscape'
import dagre from 'dagre'
import { useCallback, useMemo, useState } from 'react'
import type React from 'react'
import ReactFlow, { Background, Controls, type Edge, type Node } from 'reactflow'
import { useStreamModules } from '../../api/modules/use-stream-modules'
import { useUserPreferences } from '../../providers/user-preferences-provider'
import { DeclNode } from './DeclNode'
import { GroupNode } from './GroupNode'
import { type FTLNode, getGraphData } from './graph-utils'
import 'reactflow/dist/style.css'

interface GraphPaneProps {
  onTapped?: (item: FTLNode | null, moduleName: string | null) => void
}

const nodeTypes = {
  groupNode: GroupNode,
  declNode: DeclNode,
}

const convertToReactFlow = (elements: ElementDefinition[], nodePositions: Record<string, { x: number; y: number }>, isDarkMode: boolean) => {
  const nodes: Node[] = []
  const edges: Edge[] = []
  const moduleNodes = new Set<string>()

  // First pass: collect all module names and create nodes
  for (const el of elements) {
    if (!el.data?.id) continue

    if (el.group === 'nodes' && el.data.type !== 'groupNode') {
      const moduleName = el.data.parent || 'default'
      moduleNodes.add(moduleName)

      const node: Node = {
        id: el.data.id,
        type: 'declNode',
        position: nodePositions[el.data.id] || { x: 0, y: 0 },
        data: {
          ...el.data,
          title: el.data.label || el.data.id,
          selected: false,
          nodeType: el.data.nodeType || 'verb', // This will help us style different node types
        },
        parentNode: moduleName,
        style: {
          backgroundColor: getNodeColor(el.data.nodeType || 'verb', isDarkMode),
          zIndex: 1,
        },
      }
      nodes.push(node)
    }
  }

  // Second pass: create module group nodes
  for (const moduleName of moduleNodes) {
    nodes.push({
      id: moduleName,
      type: 'groupNode',
      position: { x: 0, y: 0 },
      data: {
        title: moduleName,
        selected: false,
        type: 'groupNode',
      },
      style: {
        padding: 20,
        backgroundColor: isDarkMode ? 'rgba(30, 41, 59, 0.5)' : 'rgba(241, 245, 249, 0.5)',
        border: `1px solid ${isDarkMode ? '#475569' : '#CBD5E1'}`,
        zIndex: -1,
      },
    })
  }

  // Add edges
  for (const el of elements) {
    if (el.group === 'edges' && el.data?.source && el.data?.target && el.data?.id) {
      edges.push({
        id: el.data.id,
        source: el.data.source,
        target: el.data.target,
        type: el.data.type === 'moduleConnection' ? 'smoothstep' : 'default',
        style: {
          stroke: isDarkMode ? '#4B5563' : '#9CA3AF',
          zIndex: 0, // Above groups (-1) but below nodes (1)
        },
      })
    }
  }

  return { nodes, edges }
}

// Helper function to get node colors based on type
const getNodeColor = (nodeType: string, isDarkMode: boolean): string => {
  const colors = {
    verb: isDarkMode ? 'rgb(99 102 241)' : 'rgb(79 70 229)', // indigo-600/500
    topic: isDarkMode ? 'rgb(168 85 247)' : 'rgb(147 51 234)', // purple-600/500
    database: isDarkMode ? 'rgb(59 130 246)' : 'rgb(37 99 235)', // blue-600/500
    config: isDarkMode ? 'rgb(14 165 233)' : 'rgb(8 145 178)', // cyan-600/500
    secret: isDarkMode ? 'rgb(234 179 8)' : 'rgb(202 138 4)', // yellow-600/500
    data: isDarkMode ? 'rgb(16 185 129)' : 'rgb(5 150 105)', // emerald-600/500
    enum: isDarkMode ? 'rgb(216 180 254)' : 'rgb(192 38 211)', // fuchsia-600/500
    default: isDarkMode ? 'rgb(99 102 241)' : 'rgb(79 70 229)', // indigo-600/500 (default)
  }
  return colors[nodeType as keyof typeof colors] || colors.default
}

// Dagre layout function
const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'LR') => {
  const dagreGraph = new dagre.graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))

  const nodeWidth = 160
  const nodeHeight = 20
  const groupPadding = 40
  const interGroupSpacing = 20 // Additional spacing between groups

  dagreGraph.setGraph({
    rankdir: direction,
    // ranker: 'network-simplex',
    nodesep: 40, // Increased from 50
    ranksep: 100, // Increased from 100
    edgesep: 20, // Increased from 20
  })

  // First pass: add all non-group nodes to dagre
  const nonGroupNodes = nodes.filter((node) => node.type !== 'groupNode')
  const groupNodes = nodes.filter((node) => node.type === 'groupNode')

  // Add virtual nodes for groups to enforce spacing
  const groupVirtualNodes = new Map<string, string[]>()

  for (const groupNode of groupNodes) {
    const childNodes = nonGroupNodes.filter((n) => n.parentNode === groupNode.id)
    if (childNodes.length === 0) continue

    // Add virtual nodes around the group's real nodes
    const virtualPrefix = `${groupNode.id}_virtual_`
    const leftNode = `${virtualPrefix}left`
    const rightNode = `${virtualPrefix}right`

    dagreGraph.setNode(leftNode, { width: 1, height: 1 })
    dagreGraph.setNode(rightNode, { width: 1, height: 1 })
    groupVirtualNodes.set(groupNode.id, [leftNode, rightNode])

    // Connect virtual nodes to enforce minimum group width
    dagreGraph.setEdge(leftNode, rightNode, { weight: 2 })
  }

  // Add all real nodes and connect them to their group's virtual nodes
  for (const node of nonGroupNodes) {
    dagreGraph.setNode(node.id, { width: nodeWidth + interGroupSpacing, height: nodeHeight })

    if (node.parentNode) {
      const virtualNodes = groupVirtualNodes.get(node.parentNode)
      if (virtualNodes) {
        const [leftNode, rightNode] = virtualNodes
        dagreGraph.setEdge(leftNode, node.id, { weight: 1 })
        dagreGraph.setEdge(node.id, rightNode, { weight: 1 })
      }
    }
  }

  // Add all edges to dagre
  for (const edge of edges) {
    dagreGraph.setEdge(edge.source, edge.target, { weight: 3 })
  }

  // Apply layout
  dagre.layout(dagreGraph)

  // Position all non-group nodes
  for (const node of nonGroupNodes) {
    const nodeWithPosition = dagreGraph.node(node.id)
    node.position = {
      x: nodeWithPosition.x - nodeWidth / 2,
      y: nodeWithPosition.y - nodeHeight / 2,
    }
  }

  // Position and size group nodes based on their children
  for (const groupNode of groupNodes) {
    const childNodes = nonGroupNodes.filter((n) => n.parentNode === groupNode.id)
    if (childNodes.length === 0) continue

    const bounds = {
      minX: Math.min(...childNodes.map((n) => n.position.x)),
      maxX: Math.max(...childNodes.map((n) => n.position.x + nodeWidth)),
      minY: Math.min(...childNodes.map((n) => n.position.y)),
      maxY: Math.max(...childNodes.map((n) => n.position.y + nodeHeight)),
    }

    groupNode.position = {
      x: bounds.minX - groupPadding,
      y: bounds.minY - groupPadding,
    }

    groupNode.style = {
      width: bounds.maxX - bounds.minX + groupPadding * 2,
      height: bounds.maxY - bounds.minY + groupPadding * 2,
    }

    // Adjust child positions to be relative to parent
    for (const child of childNodes) {
      child.position = {
        x: child.position.x - groupNode.position.x,
        y: child.position.y - groupNode.position.y,
      }
    }
  }

  return { nodes, edges }
}

export const GraphPane: React.FC<GraphPaneProps> = ({ onTapped }) => {
  const modules = useStreamModules()
  const { isDarkMode } = useUserPreferences()
  const [nodePositions] = useState<Record<string, { x: number; y: number }>>({})

  const elements = useMemo(() => {
    const cytoscapeElements = getGraphData(modules.data, isDarkMode, nodePositions)
    return convertToReactFlow(cytoscapeElements, nodePositions, isDarkMode)
  }, [modules.data, isDarkMode, nodePositions])

  const { nodes: layoutedNodes, edges: layoutedEdges } = useMemo(() => {
    if (!elements.nodes.length) return { nodes: [], edges: [] }
    return getLayoutedElements(elements.nodes, elements.edges)
  }, [elements])

  const onNodeClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      onTapped?.(node.data.item, node.id)
    },
    [onTapped],
  )

  const onPaneClick = useCallback(() => {
    onTapped?.(null, null)
  }, [onTapped])

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <ReactFlow
        nodes={layoutedNodes}
        edges={layoutedEdges}
        nodeTypes={nodeTypes}
        onNodeClick={onNodeClick}
        onPaneClick={onPaneClick}
        fitView
        minZoom={0.1}
        maxZoom={2}
        defaultEdgeOptions={{
          type: 'smoothstep',
          // Remove default edge style since we're setting it per edge
        }}
      >
        <Background />
        <Controls />
      </ReactFlow>
    </div>
  )
}
