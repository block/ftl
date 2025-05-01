import type { Edges } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { Position } from '../../../protos/xyz/block/ftl/schema/v1/schema_pb'
import type { ExpandablePanelProps } from '../../../shared/components/ExpandablePanel'
import { HoverPopup } from '../../../shared/components/HoverPopup'
import { RightPanelAttribute } from '../../../shared/components/RightPanelAttribute'
import { Schema } from '../schema/Schema'
import { EditDeclButton } from './EditDeclButton'
import { References } from './References'

export const DeclDefaultPanels = (moduleName: string, schema?: string, edges?: Edges, position?: Position) => {
  const panels = [] as ExpandablePanelProps[]

  if (position) {
    const baseFilename = position.filename.split('/').pop() ?? position.filename

    panels.push({
      title: 'File',
      expanded: true,
      children: [
        <RightPanelAttribute
          key='filename'
          name='Filename'
          value={
            <HoverPopup popupContent={position.filename}>
              <span>{baseFilename}</span>
            </HoverPopup>
          }
        />,
        <RightPanelAttribute key='line' name='Line' value={Number(position.line).toString()} />,
        <RightPanelAttribute key='column' name='Column' value={Number(position.column).toString()} />,
        <RightPanelAttribute
          key='edit'
          name='Edit'
          value={<EditDeclButton path={position.filename} line={Number(position.line)} column={Number(position.column)} />}
        />,
      ],
    })
  }

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
