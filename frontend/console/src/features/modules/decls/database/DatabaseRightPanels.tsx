import type { Database } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../../shared/components/ExpandablePanel'
import { RightPanelAttribute } from '../../../../shared/components/RightPanelAttribute'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { DeclGraphPane } from '../DeclGraphPane'

export const databasePanels = (moduleName: string, database: Database, showGraph = true) => {
  const panels: ExpandablePanelProps[] = [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={database.database?.name} />,
        <RightPanelAttribute key='type' name='Type' value={database.database?.type} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, database.schema, database.edges),
  ]

  if (showGraph) {
    panels.push({
      title: 'Graph',
      expanded: true,
      children: <DeclGraphPane declName={database.database?.name || ''} declType='database' moduleName={moduleName} edges={database.edges} />,
    })
  }

  return panels
}
