import type { Type } from '../../protos/xyz/block/ftl/schema/v1/schema_pb'

export const defaultForType = (type: Type) => {
  switch (type.value.case) {
    case 'string':
      return ''
    case 'int':
      return 0
    case 'float':
      return 0
    case 'bool':
      return false
    case 'array':
      return []
    case 'ref':
      return {}
    case 'map':
      return {}
    case 'optional':
      return null
    case 'any':
      return null
    case 'unit':
      return null
  }

  return null
}
