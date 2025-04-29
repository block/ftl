import type { Topic } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../../shared/components/ExpandablePanel'
import { RightPanelAttribute } from '../../../../shared/components/RightPanelAttribute'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { DeclGraphPane } from '../DeclGraphPane'
import { DeclLink } from '../DeclLink'

export const topicPanels = (moduleName: string, topic: Topic, showGraph = true) => {
  const panels: ExpandablePanelProps[] = [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={topic.topic?.name} />,
        <div key='event' className='flex justify-between space-x-2 items-center text-sm'>
          <span className='text-gray-500 dark:text-gray-400'>Event</span>
          <span className='flex-1 min-w-0 text-right'>
            {topic.topic?.event?.value.case === 'ref' ? (
              <DeclLink moduleName={topic.topic.event.value.value.module} declName={topic.topic.event.value.value.name} />
            ) : (
              <pre className='whitespace-pre-wrap overflow-auto'>{topic.topic?.event?.value.case}</pre>
            )}
          </span>
        </div>,
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
