import type { Config } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../../shared/components/ExpandablePanel'
import { RightPanelAttribute } from '../../../../shared/components/RightPanelAttribute'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { DeclGraphPane } from '../DeclGraphPane'

export const configPanels = (moduleName: string, config: Config, showGraph = true) => {
  const panels: ExpandablePanelProps[] = [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={config.config?.name} />,
        <RightPanelAttribute key='type' name='Type' value={config.config?.type?.value.case ?? ''} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, config.schema, config.edges),
  ]

  if (showGraph) {
    panels.push({
      title: 'Graph',
      expanded: true,
      children: <DeclGraphPane declName={config.config?.name || ''} declType='config' moduleName={moduleName} edges={config.edges} />,
    })
  }

  return panels
}
