import type { EdgeDefinition, ElementDefinition } from 'cytoscape'
import type { StreamModulesResult } from '../../api/modules/use-stream-modules'
import type { Config, Data, Database, Enum, Module, Secret, Topic, Verb } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { getNodeBackgroundColor } from './graph-styles'

export type FTLNode = Module | Verb | Secret | Config | Data | Database | Topic | Enum

const createParentNode = (module: Module, nodePositions: Record<string, { x: number; y: number }>) => ({
  group: 'nodes' as const,
  data: {
    id: module.name,
    label: module.name,
    type: 'groupNode',
    item: module,
  },
  ...(nodePositions[module.name] && {
    position: nodePositions[module.name],
  }),
})

const createChildNode = (
  parentName: string,
  childId: string,
  childLabel: string,
  childType: string,
  nodePositions: Record<string, { x: number; y: number }>,
  item: FTLNode,
  isDarkMode: boolean,
) => ({
  group: 'nodes' as const,
  data: {
    id: childId,
    label: childLabel,
    type: 'node',
    nodeType: childType,
    parent: parentName,
    item,
    backgroundColor: getNodeBackgroundColor(isDarkMode, childType),
  },
  ...(nodePositions[childId] && {
    position: nodePositions[childId],
  }),
})

const createModuleChildren = (module: Module, nodePositions: Record<string, { x: number; y: number }>, isDarkMode: boolean) => {
  const children = [
    // Create nodes for configs
    ...(module.configs || []).map((config: Config) =>
      createChildNode(module.name, nodeId(module.name, config.config?.name), config.config?.name || '', 'config', nodePositions, config, isDarkMode),
    ),
    // Create nodes for databases
    ...(module.databases || []).map((database: Database) =>
      createChildNode(
        module.name,
        nodeId(module.name, database.database?.name),
        database.database?.name || '',
        'database',
        nodePositions,
        database,
        isDarkMode,
      ),
    ),
    // Create nodes for enums
    ...(module.enums || []).map((enumDecl: Enum) =>
      createChildNode(module.name, nodeId(module.name, enumDecl.enum?.name), enumDecl.enum?.name || '', 'enum', nodePositions, enumDecl, isDarkMode),
    ),
    // Create nodes for secrets
    ...(module.secrets || []).map((secret: Secret) =>
      createChildNode(module.name, nodeId(module.name, secret.secret?.name), secret.secret?.name || '', 'secret', nodePositions, secret, isDarkMode),
    ),
    // Create nodes for topics
    ...(module.topics || []).map((topic: Topic) =>
      createChildNode(module.name, nodeId(module.name, topic.topic?.name), topic.topic?.name || '', 'topic', nodePositions, topic, isDarkMode),
    ),
    // Create nodes for verbs
    ...(module.verbs || []).map((verb: Verb) =>
      createChildNode(module.name, nodeId(module.name, verb.verb?.name), verb.verb?.name || '', 'verb', nodePositions, verb, isDarkMode),
    ),
  ]
  return children
}

const createChildEdge = (sourceModule: string, sourceVerb: string, targetModule: string, targetVerb: string) => {
  // Skip if any of the components are undefined or empty strings
  if (
    !sourceModule ||
    !sourceVerb ||
    !targetModule ||
    !targetVerb ||
    sourceModule === 'undefined' ||
    sourceVerb === 'undefined' ||
    targetModule === 'undefined' ||
    targetVerb === 'undefined'
  ) {
    return null
  }

  return {
    group: 'edges' as const,
    data: {
      id: `edge-${nodeId(sourceModule, sourceVerb)}->${nodeId(targetModule, targetVerb)}`,
      source: nodeId(sourceModule, sourceVerb),
      target: nodeId(targetModule, targetVerb),
      type: 'childConnection',
    },
  }
}

const createModuleEdge = (sourceModule: string, targetModule: string) => {
  // Skip if any of the modules are undefined or empty strings
  if (!sourceModule || !targetModule || sourceModule === 'undefined' || targetModule === 'undefined') {
    return null
  }

  return {
    group: 'edges' as const,
    data: {
      id: `module-${sourceModule}->${targetModule}`,
      source: nodeId(sourceModule),
      target: nodeId(targetModule),
      type: 'moduleConnection',
    },
  }
}

const createEdges = (modules: Module[]) => {
  const edges: EdgeDefinition[] = []
  const moduleConnections = new Set<string>() // Track unique module connections
  const existingNodes = new Set<string>() // Track all valid node IDs

  // First collect all valid node IDs
  for (const module of modules) {
    existingNodes.add(module.name)

    for (const verb of module.verbs || []) {
      if (verb.verb?.name) {
        existingNodes.add(nodeId(module.name, verb.verb.name))
      }
    }
    for (const config of module.configs || []) {
      if (config.config?.name) {
        existingNodes.add(nodeId(module.name, config.config.name))
      }
    }
    for (const secret of module.secrets || []) {
      if (secret.secret?.name) {
        existingNodes.add(nodeId(module.name, secret.secret.name))
      }
    }
    for (const database of module.databases || []) {
      if (database.database?.name) {
        existingNodes.add(nodeId(module.name, database.database.name))
      }
    }
    for (const topic of module.topics || []) {
      if (topic.topic?.name) {
        existingNodes.add(nodeId(module.name, topic.topic.name))
      }
    }
  }

  for (const module of modules) {
    // For each verb in the module
    for (const verb of module.verbs || []) {
      // For each reference in the verb
      for (const ref of verb.references || []) {
        // Skip self-referential edges
        if (ref.module === module.name && ref.name === verb.verb?.name) continue

        // Skip if source or target nodes don't exist
        const sourceId = nodeId(ref.module, ref.name)
        const targetId = nodeId(module.name, verb.verb?.name)
        if (!existingNodes.has(sourceId) || !existingNodes.has(targetId)) continue

        // Only create verb-to-verb child edges
        const edge = createChildEdge(ref.module, ref.name, module.name, verb.verb?.name || '')
        if (edge) edges.push(edge)

        // Track module-to-module connection for all reference types
        // Skip self-referential module connections
        if (ref.module !== module.name) {
          const [sourceModule, targetModule] = [module.name, ref.module].sort()
          moduleConnections.add(`${sourceModule}-${targetModule}`)
        }
      }
    }

    for (const config of module.configs || []) {
      // For each reference in the config
      for (const ref of config.references || []) {
        // Skip self-referential edges
        if (ref.module === module.name && ref.name === config.config?.name) continue

        // Skip if source or target nodes don't exist
        const sourceId = nodeId(ref.module, ref.name)
        const targetId = nodeId(module.name, config.config?.name)
        if (!existingNodes.has(sourceId) || !existingNodes.has(targetId)) continue

        const edge = createChildEdge(ref.module, ref.name, module.name, config.config?.name || '')
        if (edge) edges.push(edge)

        // Skip self-referential module connections
        if (ref.module !== module.name) {
          const [sourceModule, targetModule] = [module.name, ref.module].sort()
          moduleConnections.add(`${sourceModule}-${targetModule}`)
        }
      }
    }

    for (const secret of module.secrets || []) {
      // For each reference in the secret
      for (const ref of secret.references || []) {
        // Skip self-referential edges
        if (ref.module === module.name && ref.name === secret.secret?.name) continue

        // Skip if source or target nodes don't exist
        const sourceId = nodeId(ref.module, ref.name)
        const targetId = nodeId(module.name, secret.secret?.name)
        if (!existingNodes.has(sourceId) || !existingNodes.has(targetId)) continue

        const edge = createChildEdge(ref.module, ref.name, module.name, secret.secret?.name || '')
        if (edge) edges.push(edge)

        // Skip self-referential module connections
        if (ref.module !== module.name) {
          const [sourceModule, targetModule] = [module.name, ref.module].sort()
          moduleConnections.add(`${sourceModule}-${targetModule}`)
        }
      }
    }

    for (const database of module.databases || []) {
      // For each reference in the database
      for (const ref of database.references || []) {
        // Skip self-referential edges
        if (ref.module === module.name && ref.name === database.database?.name) continue

        // Skip if source or target nodes don't exist
        const sourceId = nodeId(ref.module, ref.name)
        const targetId = nodeId(module.name, database.database?.name)
        if (!existingNodes.has(sourceId) || !existingNodes.has(targetId)) continue

        const edge = createChildEdge(ref.module, ref.name, module.name, database.database?.name || '')
        if (edge) edges.push(edge)

        // Skip self-referential module connections
        if (ref.module !== module.name) {
          const [sourceModule, targetModule] = [module.name, ref.module].sort()
          moduleConnections.add(`${sourceModule}-${targetModule}`)
        }
      }
    }

    for (const topic of module.topics || []) {
      // For each reference in the topic
      for (const ref of topic.references || []) {
        // Skip self-referential edges
        if (ref.module === module.name && ref.name === topic.topic?.name) continue

        // Skip if source or target nodes don't exist
        const sourceId = nodeId(ref.module, ref.name)
        const targetId = nodeId(module.name, topic.topic?.name)
        if (!existingNodes.has(sourceId) || !existingNodes.has(targetId)) continue

        const edge = createChildEdge(ref.module, ref.name, module.name, topic.topic?.name || '')
        if (edge) edges.push(edge)

        // Skip self-referential module connections
        if (ref.module !== module.name) {
          const [sourceModule, targetModule] = [module.name, ref.module].sort()
          moduleConnections.add(`${sourceModule}-${targetModule}`)
        }
      }
    }
  }

  // Create module-level edges for each unique module connection
  for (const connection of moduleConnections) {
    const [sourceModule, targetModule] = connection.split('-')
    // Skip if either module doesn't exist
    if (!existingNodes.has(sourceModule) || !existingNodes.has(targetModule)) continue
    const edge = createModuleEdge(sourceModule, targetModule)
    if (edge) edges.push(edge)
  }

  return edges
}

export const getGraphData = (
  modules: StreamModulesResult | undefined,
  isDarkMode: boolean,
  nodePositions: Record<string, { x: number; y: number }> = {},
): ElementDefinition[] => {
  if (!modules) return []

  const filteredModules = modules.modules.filter((module) => module.name !== 'builtin')

  return [
    ...filteredModules.map((module) => createParentNode(module, nodePositions)),
    ...filteredModules.flatMap((module) => createModuleChildren(module, nodePositions, isDarkMode)),
    ...createEdges(filteredModules),
  ]
}

const nodeId = (moduleName: string, name?: string) => {
  if (!name) return moduleName
  return `${moduleName}.${name}`
}
