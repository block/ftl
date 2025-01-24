import { AttributeBadge } from '../../../components/AttributeBadge'
import { CodeBlockWithTitle } from '../../../components/CodeBlockWithTitle'
import type { Event, IngressEvent } from '../../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { formatDuration } from '../../../utils/date.utils'
import { DeploymentCard } from '../../deployments/DeploymentCard'
import { refString } from '../../modules/decls/verb/verb.utils'

export const TraceDetailsIngress = ({ event }: { event: Event }) => {
  const ingress = event.entry.value as IngressEvent
  return (
    <>
      <span className='text-xl font-semibold'>Call Details</span>

      {ingress.request && <CodeBlockWithTitle title='Request' code={JSON.stringify(JSON.parse(ingress.request), null, 2)} />}

      {ingress.response && <CodeBlockWithTitle title='Response' code={JSON.stringify(JSON.parse(ingress.response), null, 2)} />}

      {ingress.requestHeader && <CodeBlockWithTitle title='Request Header' code={JSON.stringify(JSON.parse(ingress.requestHeader), null, 2)} />}

      {ingress.responseHeader && <CodeBlockWithTitle title='Response Header' code={JSON.stringify(JSON.parse(ingress.responseHeader), null, 2)} />}

      {ingress.error && <CodeBlockWithTitle title='Error' code={ingress.error} />}

      <DeploymentCard className='mt-4' deploymentKey={ingress.deploymentKey} />

      <ul className='pt-4 space-y-2'>
        <li>
          <AttributeBadge name='status' value={ingress.statusCode.toString()} />
        </li>
        <li>
          <AttributeBadge name='method' value={ingress.method} />
        </li>
        <li>
          <AttributeBadge name='path' value={ingress.path} />
        </li>
        {ingress.requestKey && (
          <li>
            <AttributeBadge name='request' value={ingress.requestKey} />
          </li>
        )}
        <li>
          <AttributeBadge name='duration' value={formatDuration(ingress.duration)} />
        </li>
        {ingress.verbRef && (
          <li>
            <AttributeBadge name='verb' value={refString(ingress.verbRef)} />
          </li>
        )}
      </ul>
    </>
  )
}
