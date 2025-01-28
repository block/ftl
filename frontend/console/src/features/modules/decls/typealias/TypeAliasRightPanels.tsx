import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { TypeAlias } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../graph/ExpandablePanel'
import { DeclDefaultPanels } from '../DeclDefaultPanels'

export const typeAliasPanels = (moduleName: string, typeAlias: TypeAlias) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={typeAlias.typealias?.name} />,
        <RightPanelAttribute key='export' name='Type' value={typeAlias.typealias?.type?.value.case} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, typeAlias.schema, typeAlias.references),
  ] as ExpandablePanelProps[]
}
