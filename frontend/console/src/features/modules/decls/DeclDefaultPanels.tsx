import type { Ref } from '../../../protos/xyz/block/ftl/schema/v1/schema_pb'
import type { ExpandablePanelProps } from '../../../shared/components/ExpandablePanel'
import { Schema } from '../schema/Schema'
import { References } from './References'

export const DeclDefaultPanels = (moduleName: string, schema?: string, references?: Ref[]) => {
  const panels = [] as ExpandablePanelProps[]

  if (schema?.trim()) {
    panels.push({
      title: 'Schema',
      expanded: true,
      padding: 'p-2',
      children: <Schema schema={schema} moduleName={moduleName} />,
    })
  }

  if (references?.length) {
    panels.push({
      title: 'References',
      expanded: true,
      children: <References references={references} />,
    })
  }

  return panels
}
