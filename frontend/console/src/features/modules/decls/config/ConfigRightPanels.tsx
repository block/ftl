import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Config } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../graph/ExpandablePanel'
import { DeclDefaultPanels } from '../DeclDefaultPanels'

export const configPanels = (moduleName: string, config: Config, schema?: string) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={config.config?.name} />,
        <RightPanelAttribute key='type' name='Type' value={config.config?.type?.value.case ?? ''} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, schema, config.references),
  ] as ExpandablePanelProps[]
}
