import type { Topic } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import { ResizablePanels } from '../../../../shared/components/ResizablePanels'
import { declIcon } from '../../module.utils'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { topicPanels } from './TopicRightPanels'

export const TopicPanel = ({ topic, moduleName, declName }: { topic: Topic; moduleName: string; declName: string }) => {
  if (!topic) {
    return
  }

  const decl = topic.topic
  if (!decl) {
    return
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <PanelHeader title='Topic' declRef={`${moduleName}.${declName}`} exported={decl.export} comments={decl.comments} />
          </div>
        }
        rightPanelHeader={<RightPanelHeader Icon={declIcon('topic', decl)} title={declName} />}
        rightPanelPanels={topicPanels(moduleName, topic)}
        storageKeyPrefix='topicPanel'
      />
    </div>
  )
}
