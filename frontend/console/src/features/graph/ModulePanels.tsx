import type { Module } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../shared/components/ExpandablePanel'
import { DeclDefaultPanels } from '../modules/decls/DeclDefaultPanels'
import { Schema } from '../modules/schema/Schema'

export const modulePanels = (module: Module): ExpandablePanelProps[] => {
  const panels = []

  panels.push({
    title: 'Schema',
    expanded: true,
    children: <Schema schema={module.schema} moduleName={module.name} />,
  })

  panels.push(...DeclDefaultPanels(module.name, module.schema))

  return panels
}
