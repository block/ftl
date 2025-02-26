import { Background, BackgroundVariant, Controls, type Edge, ReactFlow as Flow, type Node, ReactFlowProvider } from '@xyflow/react'
import dagre from 'dagre'
import { useCallback, useEffect, useMemo, useState } from 'react'
import type React from 'react'
import { useUserPreferences } from '../../shared/providers/user-preferences-provider'
import { hashString } from '../../shared/utils/string.utils'
import type { StreamModulesResult } from '../modules/hooks/use-stream-modules'
import { DeclNode } from './DeclNode'
import { GroupNode } from './GroupNode'
import { type FTLNode, getGraphData } from './graph-utils'
import '@xyflow/react/dist/style.css'
import './graph.css'

const NODE_TYPES = {
  groupNode: GroupNode,
  declNode: DeclNode,
}

interface GraphPaneProps {
  modules?: StreamModulesResult
  onTapped?: (item: FTLNode | null, moduleName: string | null) => void
}

// Dagre layout function
const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'LR') => {
  const dagreGraph = new dagre.graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))

  const nodeWidth = 160
  const nodeHeight = 20
  const groupPadding = 30
  const interGroupSpacing = 80
  const groupSpacing = 10
  const verticalSpacing = 30
  const intraGroupSpacing = 25
  const titleHeight = 30

  dagreGraph.setGraph({
    rankdir: direction,
    nodesep: 10,
    ranksep: 10,
    edgesep: 10,
    marginx: 30,
    marginy: 50,
  })

  // First pass: add all non-group nodes to dagre
  const nonGroupNodes = nodes.filter((node) => node.type !== 'groupNode')
  const groupNodes = nodes.filter((node) => node.type === 'groupNode')

  // Group nodes by their parent module
  const nodesByModule = new Map<string, Node[]>()
  for (const node of nonGroupNodes) {
    if (node.parentId) {
      const nodes = nodesByModule.get(node.parentId) || []
      nodes.push(node)
      nodesByModule.set(node.parentId, nodes)
    }
  }

  // Create a map to track module ranks
  const moduleRanks = new Map<string, number>()
  let currentRank = 0

  // First, assign ranks to modules based on their connections
  for (const edge of edges) {
    const sourceModule = nonGroupNodes.find((n) => n.id === edge.source)?.parentId
    const targetModule = nonGroupNodes.find((n) => n.id === edge.target)?.parentId

    if (sourceModule && targetModule && sourceModule !== targetModule) {
      if (!moduleRanks.has(sourceModule)) {
        moduleRanks.set(sourceModule, currentRank++)
      }
      if (!moduleRanks.has(targetModule)) {
        const sourceRank = moduleRanks.get(sourceModule) ?? 0
        moduleRanks.set(targetModule, Math.max(sourceRank + 1, currentRank++))
      }
    }
  }

  // Add virtual nodes for groups to enforce spacing
  const groupVirtualNodes = new Map<string, string[]>()

  // Add virtual nodes for each group to maintain spacing
  for (const groupNode of groupNodes) {
    const childNodes = nodesByModule.get(groupNode.id) || []
    if (childNodes.length === 0) continue

    const virtualPrefix = `${groupNode.id}_virtual_`
    const leftNode = `${virtualPrefix}left`
    const rightNode = `${virtualPrefix}right`
    const topNode = `${virtualPrefix}top`

    // Add virtual nodes with both horizontal and vertical spacing
    dagreGraph.setNode(leftNode, { width: groupSpacing, height: nodeHeight })
    dagreGraph.setNode(rightNode, { width: groupSpacing, height: nodeHeight })
    dagreGraph.setNode(topNode, { width: nodeWidth, height: verticalSpacing / 2 })

    groupVirtualNodes.set(groupNode.id, [leftNode, rightNode, topNode])

    // Connect virtual nodes with weights
    dagreGraph.setEdge(leftNode, rightNode, { weight: 4 })

    // If module has a rank, use it to influence vertical positioning
    const rank = moduleRanks.get(groupNode.id)
    if (rank !== undefined) {
      dagreGraph.setNode(topNode, { rank: rank * 2 })
    }

    // Add nodes with compact intra-group spacing
    let nodeRank = 0
    for (const node of childNodes) {
      dagreGraph.setNode(node.id, {
        width: nodeWidth + interGroupSpacing,
        height: nodeHeight + intraGroupSpacing,
        rank: (rank ?? 0) * 3 + Math.floor(nodeRank++ / 2),
      })

      // Connect to virtual nodes with adjusted weights for tighter packing
      dagreGraph.setEdge(leftNode, node.id, { weight: 2 })
      dagreGraph.setEdge(node.id, rightNode, { weight: 2 })
      dagreGraph.setEdge(topNode, node.id, { weight: 1 })
    }
  }

  // Add edges with increased weight for vertical separation
  for (const edge of edges) {
    const sourceModule = nonGroupNodes.find((n) => n.id === edge.source)?.parentId
    const targetModule = nonGroupNodes.find((n) => n.id === edge.target)?.parentId

    // If edge crosses module boundaries, give it more weight
    const weight = sourceModule && targetModule && sourceModule !== targetModule ? 5 : 2
    dagreGraph.setEdge(edge.source, edge.target, { weight })
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

  // Second pass: Calculate group bounds and adjust for overlaps
  interface GroupBounds {
    id: string
    x: number
    y: number
    width: number
    height: number
    childNodes: Node[]
  }

  const groupBounds: GroupBounds[] = []

  // Calculate initial bounds for each group
  for (const groupNode of groupNodes) {
    const childNodes = nodesByModule.get(groupNode.id) || []
    if (childNodes.length === 0) continue

    const bounds = {
      minX: Math.min(...childNodes.map((n) => n.position.x)) - groupPadding,
      maxX: Math.max(...childNodes.map((n) => n.position.x + nodeWidth)) + groupPadding,
      minY: Math.min(...childNodes.map((n) => n.position.y)) - (groupPadding + titleHeight),
      maxY: Math.max(...childNodes.map((n) => n.position.y + nodeHeight)) + groupPadding,
    }

    groupBounds.push({
      id: groupNode.id,
      x: bounds.minX,
      y: bounds.minY,
      width: bounds.maxX - bounds.minX,
      height: bounds.maxY - bounds.minY,
      childNodes,
    })
  }

  // Sort groups by vertical position for overlap resolution
  groupBounds.sort((a, b) => a.y - b.y)

  // Resolve overlaps by pushing groups down
  for (let i = 1; i < groupBounds.length; i++) {
    const currentGroup = groupBounds[i]

    // Check for overlaps with all previous groups
    for (let j = 0; j < i; j++) {
      const previousGroup = groupBounds[j]

      // Check if groups overlap horizontally
      const horizontalOverlap = !(
        currentGroup.x + currentGroup.width < previousGroup.x - interGroupSpacing || currentGroup.x > previousGroup.x + previousGroup.width + interGroupSpacing
      )

      // Check if groups overlap vertically or are too close
      const verticalOverlap = !(currentGroup.y > previousGroup.y + previousGroup.height + verticalSpacing)

      // If there's both horizontal proximity and vertical overlap/proximity
      if (horizontalOverlap && verticalOverlap) {
        // Calculate the minimum shift needed to resolve overlap with extra spacing
        const verticalShift = previousGroup.y + previousGroup.height - currentGroup.y + verticalSpacing

        // Shift current group and all its children down
        currentGroup.y += verticalShift
        for (const node of currentGroup.childNodes) {
          node.position.y += verticalShift
        }

        // Also shift all subsequent groups down
        for (let k = i + 1; k < groupBounds.length; k++) {
          const nextGroup = groupBounds[k]
          nextGroup.y += verticalShift
          for (const node of nextGroup.childNodes) {
            node.position.y += verticalShift
          }
        }
      }
    }
  }

  // Update group node positions and dimensions
  for (const group of groupBounds) {
    const groupNode = groupNodes.find((n) => n.id === group.id)
    if (groupNode) {
      groupNode.position = {
        x: group.x,
        y: group.y,
      }
      groupNode.style = {
        width: group.width,
        height: group.height,
      }

      // Make child positions relative to the group
      for (const child of group.childNodes) {
        child.position = {
          x: child.position.x - group.x,
          y: child.position.y - group.y,
        }
      }
    }
  }

  return { nodes, edges }
}

export const GraphPane: React.FC<GraphPaneProps> = ({ modules, onTapped }) => {
  const { isDarkMode } = useUserPreferences()
  const [nodePositions] = useState<Record<string, { x: number; y: number }>>({})
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [moduleKey, setModuleKey] = useState<string>('empty')

  useEffect(() => {
    const updateKey = async () => {
      if (!modules?.modules) {
        setModuleKey('empty')
        return
      }
      const fullKey = modules.modules.map((m) => `${m.name}:${m.schema}`).join('-')
      const hash = await hashString(fullKey)
      setModuleKey(hash)
    }
    updateKey()
  }, [modules])

  const { nodes, edges } = useMemo(() => {
    return getGraphData(modules, isDarkMode, nodePositions, selectedNodeId)
  }, [modules, isDarkMode, nodePositions, selectedNodeId])

  const { nodes: layoutedNodes, edges: layoutedEdges } = useMemo(() => {
    if (!nodes.length) return { nodes: [], edges: [] }
    return getLayoutedElements(nodes, edges)
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
    <ReactFlowProvider>
      <div className={isDarkMode ? 'dark' : 'light'} style={{ width: '100%', height: '100%', position: 'relative' }}>
        <Flow
          key={moduleKey}
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
