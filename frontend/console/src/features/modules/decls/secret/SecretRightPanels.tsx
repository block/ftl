import { RightPanelAttribute } from '../../../../components/RightPanelAttribute'
import type { Secret } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../graph/ExpandablePanel'
import { DeclDefaultPanels } from '../DeclDefaultPanels'

export const secretPanels = (moduleName: string, secret: Secret, schema?: string) => {
  return [
    {
      title: 'Details',
      expanded: true,
      children: [
        <RightPanelAttribute key='name' name='Name' value={secret.secret?.name} />,
        <RightPanelAttribute key='type' name='Type' value={secret.secret?.type?.value.case ?? ''} />,
      ],
    },
    ...DeclDefaultPanels(moduleName, schema, secret.references),
  ] as ExpandablePanelProps[]
}
