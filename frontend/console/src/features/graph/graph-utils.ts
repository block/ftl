import type { Edge, Node } from '@xyflow/react'
import * as dagre from 'dagre'
import { Config, Data, Database, type Enum, Module, Secret, Topic, Verb } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { type Metadata, type Ref, Visibility } from '../../protos/xyz/block/ftl/schema/v1/schema_pb'
import type { ExpandablePanelProps } from '../../shared/components/ExpandablePanel'
import { configPanels } from '../modules/decls/config/ConfigRightPanels'
import { dataPanels } from '../modules/decls/data/DataRightPanels'
import { databasePanels } from '../modules/decls/database/DatabaseRightPanels'
import { secretPanels } from '../modules/decls/secret/SecretRightPanels'
import { topicPanels } from '../modules/decls/topic/TopicRightPanels'
import { verbPanels } from '../modules/decls/verb/VerbRightPanel'
import type { StreamModulesResult } from '../modules/hooks/use-stream-modules'
import { getVerbType } from '../modules/module.utils'
import { modulePanels } from './ModulePanels'
import { getNodeBackgroundColor } from './graph-styles'

export type FTLNode = Module | Verb | Secret | Config | Data | Database | Topic | Enum

interface GraphData {
  nodes: Node[]
  edges: Edge[]
}

export const createNode = (
  id: string,
  label: string,
  type: 'groupNode' | 'declNode',
  nodeType: string,
  position: { x: number; y: number } | undefined,
  parentId: string | undefined,
  item: FTLNode,
  isDarkMode: boolean,
  isSelected: boolean,
): Node => ({
  id,
  type,
  position: position || { x: 0, y: 0 },
  ...(parentId && { parentId }),
  data: {
    title: label,
    selected: isSelected,
    type,
    nodeType,
    item,
    style: {
      backgroundColor: getNodeBackgroundColor(isDarkMode, nodeType),
    },
    zIndex: type === 'groupNode' ? -1 : 2,
    isExported: type === 'declNode' ? getIsExported(item) : undefined,
  },
})

// Helper function to determine if a declaration is exported
export const getIsExported = (item: FTLNode): boolean | undefined => {
  if ('verb' in item && item.verb && typeof item.verb === 'object' && 'export' in item.verb) {
    return !!item.verb.export
  }
  if ('data' in item && item.data && typeof item.data === 'object' && 'export' in item.data) {
    return !!item.data.export
  }
  if ('typealias' in item && item.typealias && typeof item.typealias === 'object' && 'export' in item.typealias) {
    return !!item.typealias.export
  }
  if ('topic' in item && item.topic && typeof item.topic === 'object' && 'export' in item.topic) {
    return !!item.topic.export
  }
  if ('config' in item && item.config && typeof item.config === 'object' && 'export' in item.config) {
    return !!item.config.export
  }
  if ('database' in item && item.database && typeof item.database === 'object' && 'export' in item.database) {
    return !!item.database.export
  }
  return undefined
}

export const createEdge = (
  sourceModule: string,
  sourceVerb: string | undefined,
  targetModule: string,
  targetVerb: string | undefined,
  isDarkMode: boolean,
  selectedNodeId: string | null,
): Edge | null => {
  // Skip if any required components are missing
  if (!sourceModule || !targetModule || !sourceVerb || !targetVerb) {
    return null
  }

  const source = nodeId(sourceModule, sourceVerb)
  const target = nodeId(targetModule, targetVerb)

  // Skip if either source or target would be undefined
  if (!source || !target) {
    return null
  }

  const id = `edge-${source}->${target}`
  const isConnectedToSelectedNode = Boolean(selectedNodeId && (source === selectedNodeId || target === selectedNodeId))

  return {
    id,
    source,
    target,
    type: 'default',
    animated: isConnectedToSelectedNode,
    zIndex: 1,
    style: {
      stroke: isConnectedToSelectedNode ? (isDarkMode ? '#EC4899' : '#F472B6') : isDarkMode ? '#4B5563' : '#9CA3AF',
      strokeWidth: isConnectedToSelectedNode ? 2 : 1,
    },
  }
}

export const getGraphData = (
  modules: StreamModulesResult | undefined,
  isDarkMode: boolean,
  nodePositions: Record<string, { x: number; y: number }> = {},
  selectedNodeId: string | null = null,
): GraphData => {
  if (!modules) return { nodes: [], edges: [] }

  const nodes: Node[] = []
  const edges: Edge[] = []
  const existingNodes = new Set<string>()
  const filteredModules = modules.modules.filter((module) => module.name !== 'builtin')

  // First pass: Create all nodes and collect valid node IDs
  for (const module of filteredModules) {
    existingNodes.add(module.name)

    // Create module (group) node
    nodes.push(
      createNode(module.name, module.name, 'groupNode', 'groupNode', nodePositions[module.name], undefined, module, isDarkMode, module.name === selectedNodeId),
    )

    // Create child nodes
    const createChildren = <T extends FTLNode>(items: T[], type: string, getName: (item: T) => string) => {
      for (const item of items || []) {
        const name = getName(item)
        if (name) {
          const id = nodeId(module.name, name)
          existingNodes.add(id)

          // For verbs, check if there's a specific sub-type
          let nodeType = type
          if (type === 'verb' && item instanceof Verb) {
            const verbType = getVerbType(item)
            if (verbType) {
              nodeType = verbType
            }
          }

          nodes.push(createNode(id, name, 'declNode', nodeType, nodePositions[id], module.name, item, isDarkMode, id === selectedNodeId))
        }
      }
    }

    createChildren(module.verbs, 'verb', (item: Verb) => item.verb?.name || '')
    createChildren(module.configs, 'config', (item: Config) => item.config?.name || '')
    createChildren(module.secrets, 'secret', (item: Secret) => item.secret?.name || '')
    createChildren(module.databases, 'database', (item: Database) => item.database?.name || '')
    createChildren(module.topics, 'topic', (item: Topic) => item.topic?.name || '')
  }

  // Remove metadata-based edge creation for publisher, subscriber, and calls
  // Process edges for all node types (including verbs)
  const processReferences = <T extends FTLNode & { edges?: { in: Array<{ module: string; name: string }>; out: Array<{ module: string; name: string }> } }>(
    module: Module,
    items: T[],
    getName: (item: T) => string,
  ) => {
    for (const item of items || []) {
      const itemName = getName(item)
      if (!itemName || !item.edges) continue

      // Process inbound edges (unchanged)
      for (const ref of item.edges.in) {
        if (!ref.name || !ref.module) continue
        if (ref.module === module.name && ref.name === itemName) continue
        const sourceId = nodeId(ref.module, ref.name)
        const targetId = nodeId(module.name, itemName)
        if (!existingNodes.has(sourceId) || !existingNodes.has(targetId)) continue

        // Special case: If this is a topic and the reference is a verb (subscriber), reverse the direction
        if (item instanceof Topic) {
          const sourceModule = filteredModules.find((m) => m.name === ref.module)
          const sourceVerb = sourceModule?.verbs.find((v) => (v.verb?.name || '') === ref.name)
          if (sourceVerb && isSubscriber(sourceVerb, module.name, itemName)) {
            const edge = createEdge(module.name, itemName, ref.module, ref.name, isDarkMode, selectedNodeId)
            if (edge) edges.push(edge)
            continue
          }
        }
        const edge = createEdge(ref.module, ref.name, module.name, itemName, isDarkMode, selectedNodeId)
        if (edge) edges.push(edge)
      }

      // Unified outbound edge processing for all node types
      for (const ref of item.edges.out) {
        if (!ref.name || !ref.module) continue
        if (ref.module === module.name && ref.name === itemName) continue
        const sourceId = nodeId(module.name, itemName)
        const targetId = nodeId(ref.module, ref.name)
        if (!existingNodes.has(sourceId) || !existingNodes.has(targetId)) continue

        const targetModule = filteredModules.find((m) => m.name === ref.module)
        const targetTopic = targetModule?.topics.find((t) => (t.topic?.name || '') === ref.name)

        // For verbs and topics, add both subscriber and publisher edges if both relationships exist
        if (targetTopic && item instanceof Verb) {
          if (isSubscriber(item, ref.module, ref.name)) {
            const edge = createEdge(ref.module, ref.name, module.name, itemName, isDarkMode, selectedNodeId)
            if (edge) edges.push(edge)
          }
          if (isPublisher(item, ref.module, ref.name)) {
            const edge = createEdge(module.name, itemName, ref.module, ref.name, isDarkMode, selectedNodeId)
            if (edge) edges.push(edge)
          }
          // Don't add the default edge for topic/verb relationships
          continue
        }
        // Default edge for other types
        const edge = createEdge(module.name, itemName, ref.module, ref.name, isDarkMode, selectedNodeId)
        if (edge) edges.push(edge)
      }
    }
  }

  for (const module of filteredModules) {
    processReferences(module, module.verbs, (item: Verb) => item.verb?.name || '')
    processReferences(module, module.configs, (item: Config) => item.config?.name || '')
    processReferences(module, module.secrets, (item: Secret) => item.secret?.name || '')
    processReferences(module, module.databases, (item: Database) => item.database?.name || '')
    processReferences(module, module.topics, (item: Topic) => item.topic?.name || '')
  }

  // Deduplicate edges
  const uniqueEdges = new Map<string, Edge>()
  for (const edge of edges) {
    const key = `${edge.source}->${edge.target}`
    uniqueEdges.set(key, edge)
  }

  return { nodes, edges: Array.from(uniqueEdges.values()) }
}

export const nodeId = (moduleName: string, name?: string) => {
  if (!name) return moduleName
  return `${moduleName}.${name}`
}

// Layout function for module-specific graphs (no groups)
export const getModuleLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'LR') => {
  const dagreGraph = new dagre.graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))

  const nodeWidth = 160
  const nodeHeight = 36

  dagreGraph.setGraph({
    rankdir: direction,
    nodesep: 50,
    ranksep: 80,
    marginx: 30,
    marginy: 30,
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

export const panelsForNode = (node: FTLNode | null, ref: string | null) => {
  if (node instanceof Module) {
    return modulePanels(node)
  }

  // If no module name is provided, we can't show the panels
  if (!ref) {
    return [] as ExpandablePanelProps[]
  }

  if (node instanceof Config) {
    return configPanels(ref, node, false)
  }
  if (node instanceof Secret) {
    return secretPanels(ref, node, false)
  }
  if (node instanceof Database) {
    return databasePanels(ref, node, false)
  }
  if (node instanceof Data) {
    return dataPanels(ref, node, false)
  }
  if (node instanceof Topic) {
    return topicPanels(ref, node, false)
  }
  if (node instanceof Verb) {
    return verbPanels(ref, node, false)
  }
  return [] as ExpandablePanelProps[]
}

function isSubscriber(verb: Verb, module: string, name: string): boolean {
  return (verb.verb?.metadata ?? []).some(
    (meta: Metadata) => meta.value.case === 'subscriber' && meta.value.value.topic?.module === module && meta.value.value.topic?.name === name,
  )
}

function isPublisher(verb: Verb, module: string, name: string): boolean {
  return (verb.verb?.metadata ?? []).some(
    (meta: Metadata) => meta.value.case === 'publisher' && (meta.value.value.topics ?? []).some((topic: Ref) => topic.module === module && topic.name === name),
  )
}

type WithVisibility = { visibility?: Visibility }
export const nodeIsExported = (node: FTLNode | undefined) => {
  if (!node) return false

  let visibility: Visibility | undefined
  if (node instanceof Config) {
    visibility = (node.config as WithVisibility).visibility
  }
  if (node instanceof Secret) {
    visibility = (node.secret as WithVisibility).visibility
  }
  if (node instanceof Database) {
    visibility = (node.database as WithVisibility).visibility
  }
  if (node instanceof Data) {
    visibility = (node.data as WithVisibility).visibility
  }
  if (node instanceof Topic) {
    visibility = (node.topic as WithVisibility).visibility
  }
  if (node instanceof Verb) {
    visibility = (node.verb as WithVisibility).visibility
  }
  return visibility === Visibility.SCOPE_MODULE || visibility === Visibility.SCOPE_REALM
}
