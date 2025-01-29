import type { TypeAlias } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../../shared/components/ExpandablePanel'
import { RightPanelAttribute } from '../../../../shared/components/RightPanelAttribute'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { DeclGraphPane } from '../DeclGraphPane'

export const typeAliasPanels = (moduleName: string, typeAlias: TypeAlias, showGraph = true) => {
  const panels: ExpandablePanelProps[] = [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={typeAlias.typealias?.name} />,
        <RightPanelAttribute key='export' name='Type' value={typeAlias.typealias?.type?.value.case} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, typeAlias.schema, typeAlias.edges),
  ]

  if (showGraph) {
    panels.push({
      title: 'Graph',
      expanded: true,
      children: <DeclGraphPane declName={typeAlias.typealias?.name || ''} declType='typealias' moduleName={moduleName} edges={typeAlias.edges} />,
    })
  }

  return panels
}
