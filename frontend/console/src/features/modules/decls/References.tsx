import { ArrowLeft05Icon, ArrowRight05Icon } from 'hugeicons-react'
import type { Edges } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { DeclLink } from './DeclLink'

export const References = ({ edges }: { edges: Edges }) => {
  return (
    <div className='text-sm space-y-4'>
      <div>
        <div className='flex items-center gap-1 text-xs text-gray-500 mb-1'>
          <ArrowRight05Icon name='arrow-down-left' className='w-4 h-4' /> Inbound References
        </div>
        <div className='space-y-1'>
          {edges.in.length === 0 ? (
            <div className='text-gray-500 text-xs'>None</div>
          ) : (
            edges.in.map((r, i) => (
              <div key={i} className='font-mono text-xs'>
                <DeclLink moduleName={r.module} declName={r.name} />
              </div>
            ))
          )}
        </div>
      </div>

      <div>
        <div className='flex items-center gap-1 text-xs text-gray-500 mb-1'>
          <ArrowLeft05Icon name='arrow-up-right' className='w-4 h-4' /> Outbound References
        </div>
        <div className='space-y-1'>
          {edges.out.length === 0 ? (
            <div className='text-gray-500 text-xs'>None</div>
          ) : (
            edges.out.map((r, i) => (
              <div key={i} className='font-mono text-xs'>
                <DeclLink moduleName={r.module} declName={r.name} />
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
