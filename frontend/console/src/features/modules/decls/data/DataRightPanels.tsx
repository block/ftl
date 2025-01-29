import type { Data } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../../shared/components/ExpandablePanel'
import { RightPanelAttribute } from '../../../../shared/components/RightPanelAttribute'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { DeclGraphPane } from '../DeclGraphPane'

export const dataPanels = (moduleName: string, data: Data, showGraph = true) => {
  const panels: ExpandablePanelProps[] = [
    {
      title: 'Details',
      expanded: true,
      children: [<RightPanelAttribute key='name' name='Name' value={data.data?.name} />],
    },
    ...DeclDefaultPanels(moduleName, data.schema, data.edges),
  ]

  if (showGraph) {
    panels.push({
      title: 'Graph',
      expanded: true,
      children: <DeclGraphPane declName={data.data?.name || ''} declType='data' moduleName={moduleName} edges={data.edges} />,
    })
  }

  return panels
}
