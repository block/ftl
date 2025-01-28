import type { Enum } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../../shared/components/ExpandablePanel'
import { RightPanelAttribute } from '../../../../shared/components/RightPanelAttribute'
import { DeclDefaultPanels } from '../DeclDefaultPanels'

export const enumPanels = (moduleName: string, enumDecl: Enum) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={enumDecl.enum?.name} />,
        <RightPanelAttribute key='type' name='Type' value={enumDecl.enum?.type?.value.case} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, enumDecl.schema, enumDecl.references),
  ] as ExpandablePanelProps[]
}
