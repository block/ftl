import type { Enum } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../../shared/components/ExpandablePanel'
import { RightPanelAttribute } from '../../../../shared/components/RightPanelAttribute'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { DeclGraphPane } from '../DeclGraphPane'

export const enumPanels = (moduleName: string, enumDecl: Enum, showGraph = true) => {
  const panels: ExpandablePanelProps[] = [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={enumDecl.enum?.name} />,
        <RightPanelAttribute key='type' name='Type' value={enumDecl.enum?.type?.value.case} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, enumDecl.schema, enumDecl.edges),
  ]

  if (showGraph) {
    panels.push({
      title: 'Graph',
      expanded: true,
      children: <DeclGraphPane declName={enumDecl.enum?.name || ''} declType='enum' moduleName={moduleName} edges={enumDecl.edges} />,
    })
  }

  return panels
}
