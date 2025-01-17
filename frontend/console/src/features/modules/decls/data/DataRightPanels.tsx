import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Data } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../graph/ExpandablePanel'
import { DeclDefaultPanels } from '../DeclDefaultPanels'

export const dataPanels = (moduleName: string, data: Data) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [<RightPanelAttribute key='name' name='Name' value={data.data?.name} />],
    },
    ...DeclDefaultPanels(moduleName, data.schema, data.references),
  ] as ExpandablePanelProps[]
}
