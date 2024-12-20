import type { PubSubPublishEvent } from '../../protos/xyz/block/ftl/timeline/v1/event_pb'
import { refString } from '../modules/decls/verb/verb.utils'

export const TimelinePubSubPublish = ({ pubSubPublish }: { pubSubPublish: PubSubPublishEvent }) => {
  const title = `${pubSubPublish.verbRef?.module ? `${refString(pubSubPublish.verbRef)} -> ` : ''} topic ${pubSubPublish.topic}`
  return (
    <span title={title}>
      {pubSubPublish.verbRef?.module && (
        <>
          <span className='text-indigo-500 dark:text-indigo-300'>{refString(pubSubPublish.verbRef)}</span>
          {' published to topic '}
        </>
      )}
      <span className='text-indigo-500 dark:text-indigo-300'>${pubSubPublish.topic}</span>
    </span>
  )
}
