import type { Module } from '../../../protos/xyz/block/ftl/console/v1/console_pb'

export const commentPrefix = '//'

export const staticKeywords = ['module', 'export']

export const declTypes = ['config', 'data', 'database', 'enum', 'topic', 'typealias', 'secret', 'subscription', 'verb']

export const declTypeMultiselectOpts = [
  {
    key: 'config',
    displayName: 'Config',
  },
  {
    key: 'data',
    displayName: 'Data',
  },
  {
    key: 'database',
    displayName: 'Database',
  },
  {
    key: 'enum',
    displayName: 'Enum',
  },
  {
    key: 'topic',
    displayName: 'Topic',
  },
  {
    key: 'typealias',
    displayName: 'Type Alias',
  },
  {
    key: 'secret',
    displayName: 'Secret',
  },
  {
    key: 'subscription',
    displayName: 'Subscription',
  },
  {
    group: 'Verb',
    key: 'cronjob',
    displayName: 'Cron Job',
  },
  {
    group: 'Verb',
    key: 'ingress',
    displayName: 'Ingress Verb',
  },
  {
    group: 'Verb',
    key: 'subscriber',
    displayName: 'Subscriber',
  },
  {
    group: 'Verb',
    key: 'sqlquery',
    displayName: 'SQL Query',
  },
  {
    group: 'Verb',
    key: 'verb',
    displayName: 'All Other Verbs',
  },
]

// Keep these in sync with common/schema/module.go#L86-L95
const skipNewLineDeclTypes = ['config', 'secret', 'database', 'topic', 'subscription', 'typealias']
const skipGapAfterTypes: { [key: string]: string[] } = {
  secret: ['config'],
  subscription: ['topic'],
}

export const specialChars = ['{', '}', '=']

export const shouldAddLeadingSpace = (lines: string[], i: number): boolean => {
  if (!isFirstLineOfBlock(lines, i)) {
    return false
  }

  for (const j in skipNewLineDeclTypes) {
    if (declTypeAndPriorLineMatch(lines, i, skipNewLineDeclTypes[j], skipNewLineDeclTypes[j])) {
      return false
    }
  }

  for (const declType in skipGapAfterTypes) {
    for (const j in skipGapAfterTypes[declType]) {
      if (declTypeAndPriorLineMatch(lines, i, declType, skipGapAfterTypes[declType][j])) {
        return false
      }
    }
  }

  return true
}

const declTypeAndPriorLineMatch = (lines: string[], i: number, declType: string, priorDeclType: string): boolean => {
  if (i === 0 || lines.length === 1) {
    return false
  }
  return regexForDeclType(declType).exec(lines[i]) !== null && regexForDeclType(priorDeclType).exec(lines[i - 1]) !== null
}

const regexForDeclType = (declType: string) => {
  return new RegExp(`^  (export )?${declType} \\w+`)
}

const isFirstLineOfBlock = (lines: string[], i: number): boolean => {
  if (i === 0) {
    // Never add space for the first block
    return false
  }
  if (lines[i].startsWith('    ')) {
    // Never add space for nested lines
    return false
  }
  if (lines[i - 1].trim().startsWith(commentPrefix)) {
    // Prior line is a comment
    return false
  }
  if (lines[i].trim().startsWith(commentPrefix)) {
    return true
  }
  const tokens = lines[i].trim().split(' ')
  if (!tokens || tokens.length === 0) {
    return false
  }
  return staticKeywords.includes(tokens[0]) || declTypes.includes(tokens[0])
}

export interface DeclSchema {
  schema: string
  declType: string
}

export const declSchemaFromModules = (moduleName: string, declName: string, modules: Module[]) => {
  // First try to find in the specified module
  const module = modules.find((module) => module.name === moduleName)
  if (module?.schema) {
    const decl = declFromModuleSchemaString(declName, module.schema)
    if (decl) {
      return decl
    }
  }

  // If not found, search all other modules
  for (const otherModule of modules) {
    if (otherModule.name === moduleName) continue // Skip the module we already checked
    if (!otherModule.schema) continue

    const decl = declFromModuleSchemaString(declName, otherModule.schema)
    if (decl) {
      return decl
    }
  }

  return undefined
}

const declFromModuleSchemaString = (declName: string, schema: string) => {
  const lines = schema.split('\n')
  const foundIdx = findDeclLinkIdx(declName, lines)

  if (foundIdx === -1) {
    return
  }

  const line = lines[foundIdx]
  let out = line
  let subLineIdx = foundIdx + 1
  while (subLineIdx < lines.length && lines[subLineIdx].startsWith('    ')) {
    out += `\n${lines[subLineIdx]}`
    subLineIdx++
  }
  // Check for closing parens
  if (subLineIdx < lines.length && line.endsWith('{') && lines[subLineIdx] === '  }') {
    out += '\n  }'
  }

  // Scan backwards for comments
  subLineIdx = foundIdx - 1
  while (subLineIdx >= 0 && lines[subLineIdx].trim().startsWith(commentPrefix)) {
    out = `${lines[subLineIdx]}\n${out}`
    subLineIdx--
  }

  const regexExecd = new RegExp(` (\\w+) ${declName}`).exec(line)
  const declType = lines[foundIdx].includes('database postgres') ? 'database' : regexExecd ? regexExecd[1] : ''
  return {
    schema: out,
    declType,
  }
}

const findDeclLinkIdx = (declName: string, lines: string[]) => {
  const regex = new RegExp(`^  (export )?\\w+ ${declName}[ (<:]`)
  const foundIdx = lines.findIndex((line) => line.match(regex))
  if (foundIdx !== -1) {
    return foundIdx
  }

  // Check for databases, for which the DB type prefaces the name.
  const dbRegex = new RegExp(`^  (export )?database postgres ${declName}`)
  return lines.findIndex((line) => line.match(dbRegex))
}
