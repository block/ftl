import type { Secret } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../../shared/components/ExpandablePanel'
import { RightPanelAttribute } from '../../../../shared/components/RightPanelAttribute'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { DeclGraphPane } from '../DeclGraphPane'

export const secretPanels = (moduleName: string, secret: Secret, showGraph = true) => {
  const panels: ExpandablePanelProps[] = [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={secret.secret?.name} />,
        <RightPanelAttribute key='type' name='Type' value={secret.secret?.type?.value.case ?? ''} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, secret.schema, secret.edges),
  ]

  if (showGraph) {
    panels.push({
      title: 'Graph',
      expanded: true,
      children: <DeclGraphPane declName={secret.secret?.name || ''} declType='secret' moduleName={moduleName} edges={secret.edges} />,
    })
  }

  return panels
}
