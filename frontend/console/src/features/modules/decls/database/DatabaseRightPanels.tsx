import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Database } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../graph/ExpandablePanel'
import { DeclDefaultPanels } from '../DeclDefaultPanels'

export const databasePanels = (moduleName: string, database: Database) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={database.database?.name} />,
        <RightPanelAttribute key='type' name='Type' value={database.database?.type} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, database.schema, database.references),
  ] as ExpandablePanelProps[]
}
