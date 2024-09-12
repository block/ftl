import type { Optional, Ref, Array as SchArray, Map as SchMap, Type } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { DeclLink } from './DeclLink'

// TypeParams ironically is built to work with the `Type` message, not the
// `TypeParameter` message, which just has a simple string param name without
// any higher level type information.
const TypeParams = ({ types }: { types?: (Type | undefined)[] }) => {
  const definedTypes = types?.filter((t) => t !== undefined)
  if (!definedTypes || definedTypes.length === 0) {
    return
  }
  return (
    <span>
      <span>{'<'}</span>
      {definedTypes.map((t, i) => [<TypeEl key='t' t={t} />, i === definedTypes.length - 1 ? '' : ', '])}
      <span>{'>'}</span>
    </span>
  )
}

export const RefLink = ({ r }: { r: Ref }) => {
  return (
    <span>
      <DeclLink moduleName={r.module} declName={r.name} />
      <TypeParams types={r.typeParameters} />
    </span>
  )
}

export const TypeEl = ({ t }: { t?: Type }) => {
  if (!t) {
    return ''
  }

  const v = t.value.value
  if (!v) {
    return t.value.case
  }

  switch (t.value.case) {
    case 'array':
      return (
        <span>
          array
          <TypeParams types={[(v as SchArray).element]} />
        </span>
      )
    case 'map':
      return (
        <span>
          map
          <TypeParams types={[(v as SchMap).key, (v as SchMap).value]} />
        </span>
      )
    case 'optional':
      return (
        <span>
          optional
          <TypeParams types={[(v as Optional).type]} />
        </span>
      )
    case 'ref':
      return <RefLink r={v as Ref} />
    default:
      return t.value.case || ''
  }
}
