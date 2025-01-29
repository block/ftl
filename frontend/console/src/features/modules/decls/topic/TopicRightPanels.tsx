import type { Topic } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../../shared/components/ExpandablePanel'
import { RightPanelAttribute } from '../../../../shared/components/RightPanelAttribute'
import { DeclDefaultPanels } from '../DeclDefaultPanels'

export const topicPanels = (moduleName: string, topic: Topic) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={topic.topic?.name} />,
        <RightPanelAttribute key='export' name='Event' value={topic.topic?.event?.value.case} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, topic.schema, topic.edges),
  ] as ExpandablePanelProps[]
}
