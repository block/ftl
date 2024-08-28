import type { Data } from '../../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { ExportBadge } from './ExportBadge'

export const DataPanel = ({ value, moduleName, declName }: { value: Data; moduleName: string; declName: string }) => {
  return (
    <div className='flex-1 py-2 px-4'>
      {value.export ? <ExportBadge /> : []}
      <div className='inline-block mr-3 align-middle'>
        <p>
          data: {moduleName}.{declName}
        </p>
        {value.comments.length > 0 ? <p className='text-xs my-1'>{value.comments}</p> : []}
      </div>
    </div>
  )
}
