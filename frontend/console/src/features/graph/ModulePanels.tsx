import type { Module } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../shared/components/ExpandablePanel'
import { DeclDefaultPanels } from '../modules/decls/DeclDefaultPanels'

export const modulePanels = (module: Module): ExpandablePanelProps[] => {
  const panels = []

  panels.push(...DeclDefaultPanels(module.name, module.schema))

  return panels
}
