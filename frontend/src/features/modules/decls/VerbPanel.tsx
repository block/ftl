import { ExportBadge } from './ExportBadge'
import type { Ref, Type, Verb } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'

const DataRef = ({ heading, r }: { heading: string, r: Ref }) => {
  return (
    <div
      className={ioBlockClassName}
      onClick={() => window.location.pathname = `/modules/${r.module}/data/${r.name}`}
    >
      <div className='text-sm'>{heading}</div>
      <span className='text-xs'>{r.module}.{r.name}</span>
    </div>
  )
}

const ioBlockClassName = 'rounded-md inline-block align-middle w-40 bg-gray-200 my-3 mr-3 py-1 px-2 hover:bg-gray-100 hover:cursor-pointer'
const IOBlock = ({ heading, t }: { heading: string, t?: Type }) => {
  if (!t) {
    return []
  }
  if (t.value.case === 'ref') {
    return <DataRef heading={heading} r={t.value.value} />
  }
  return (
    <div className={ioBlockClassName}>
      <div className='text-sm'>{heading}</div>
      <div className='text-xs'>{t.value.case}</div>
    </div>
  )
}

export const VerbPanel = ({ value, moduleName, declName }: { value: Verb, moduleName: string, declName: string }) => {
  return (
    <div className='flex-1 py-2 px-4'>
      {value.export ? <ExportBadge /> : []}
      <div className='inline-block mr-3 align-middle'>
        <p>verb: {moduleName}.{declName}</p>
        {value.comments.length > 0 ? <p className='text-xs my-1'>{value.comments}</p> : []}
      </div>
      <IOBlock heading='Request' t={value.request} />
      <IOBlock heading='Response' t={value.response} />
    </div>
  )
}
