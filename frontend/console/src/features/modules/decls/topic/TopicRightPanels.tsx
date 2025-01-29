import type { Topic } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../../shared/components/ExpandablePanel'
import { RightPanelAttribute } from '../../../../shared/components/RightPanelAttribute'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { DeclGraphPane } from '../DeclGraphPane'

export const topicPanels = (moduleName: string, topic: Topic, showGraph = true) => {
  const panels: ExpandablePanelProps[] = [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={topic.topic?.name} />,
        <RightPanelAttribute key='export' name='Event' value={topic.topic?.event?.value.case} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, topic.schema, topic.edges),
  ]

  if (showGraph) {
    panels.push({
      title: 'Graph',
      expanded: true,
      children: <DeclGraphPane declName={topic.topic?.name || ''} declType='topic' moduleName={moduleName} edges={topic.edges} />,
    })
  }

  return panels
}
