import type { Edges } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../shared/components/ExpandablePanel'
import { Schema } from '../schema/Schema'
import { References } from './References'

export const DeclDefaultPanels = (moduleName: string, schema?: string, edges?: Edges) => {
  const panels = [] as ExpandablePanelProps[]

  if (edges?.in.length || edges?.out.length) {
    panels.push({
      title: 'References',
      expanded: true,
      children: <References edges={edges} />,
    })
  }

  if (schema?.trim()) {
    panels.push({
      title: 'Schema',
      expanded: true,
      padding: 'p-2',
      children: <Schema schema={schema} moduleName={moduleName} />,
    })
  }

  return panels
}
