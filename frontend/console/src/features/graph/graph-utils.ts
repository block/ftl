import type { Edge, Node } from 'reactflow'
import type { Config, Data, Database, Enum, Module, Secret, Topic, Verb } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import type { StreamModulesResult } from '../modules/hooks/use-stream-modules'
import { getNodeBackgroundColor } from './graph-styles'

export type FTLNode = Module | Verb | Secret | Config | Data | Database | Topic | Enum

interface GraphData {
  nodes: Node[]
  edges: Edge[]
}

const createNode = (
  id: string,
  label: string,
  type: 'groupNode' | 'declNode',
  nodeType: string,
  position: { x: number; y: number } | undefined,
  parentNode: string | undefined,
  item: FTLNode,
  isDarkMode: boolean,
  isSelected: boolean,
): Node => ({
  id,
  type,
  position: position || { x: 0, y: 0 },
  ...(parentNode && { parentNode }),
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
  },
})

const createEdge = (
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
          nodes.push(createNode(id, name, 'declNode', type, nodePositions[id], module.name, item, isDarkMode, id === selectedNodeId))
        }
      }
    }

    createChildren(module.verbs, 'verb', (item: Verb) => item.verb?.name || '')
    createChildren(module.configs, 'config', (item: Config) => item.config?.name || '')
    createChildren(module.secrets, 'secret', (item: Secret) => item.secret?.name || '')
    createChildren(module.databases, 'database', (item: Database) => item.database?.name || '')
    createChildren(module.topics, 'topic', (item: Topic) => item.topic?.name || '')
    createChildren(module.enums, 'enum', (item: Enum) => item.enum?.name || '')
  }

  // Second pass: Create edges
  const processReferences = <T extends FTLNode & { references?: Array<{ module: string; name: string }> }>(
    module: Module,
    items: T[],
    getName: (item: T) => string,
  ) => {
    for (const item of items || []) {
      const itemName = getName(item)
      // Skip if the item name is empty
      if (!itemName || itemName === '') continue
      if (!item.references) continue

      for (const ref of item.references) {
        // Skip if reference name is empty
        if (!ref.name || ref.name === '') continue
        // Skip if reference module is empty
        if (!ref.module || ref.module === '') continue

        // Skip self-referential edges
        if (ref.module === module.name && ref.name === itemName) continue

        // Skip if source or target nodes don't exist
        const sourceId = nodeId(ref.module, ref.name)
        const targetId = nodeId(module.name, itemName)
        if (!existingNodes.has(sourceId) || !existingNodes.has(targetId)) continue

        const edge = createEdge(ref.module, ref.name, module.name, itemName, isDarkMode, selectedNodeId)
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

  return { nodes, edges }
}

const nodeId = (moduleName: string, name?: string) => {
  if (!name) return moduleName
  return `${moduleName}.${name}`
}
