import { useState } from 'react'
import { Config, Data, Database, Enum, Module, Secret, Topic, Verb } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../shared/components/ExpandablePanel'
import { Loader } from '../../shared/components/Loader'
import { ResizablePanels } from '../../shared/components/ResizablePanels'
import { configPanels } from '../modules/decls/config/ConfigRightPanels'
import { dataPanels } from '../modules/decls/data/DataRightPanels'
import { databasePanels } from '../modules/decls/database/DatabaseRightPanels'
import { enumPanels } from '../modules/decls/enum/EnumRightPanels'
import { secretPanels } from '../modules/decls/secret/SecretRightPanels'
import { topicPanels } from '../modules/decls/topic/TopicRightPanels'
import { verbPanels } from '../modules/decls/verb/VerbRightPanel'
import { useModules } from '../modules/hooks/use-modules'
import { Timeline } from '../timeline/Timeline'
import { GraphPane } from './GraphPane'
import { modulePanels } from './ModulePanels'
import { headerForNode } from './RightPanelHeader'
import type { FTLNode } from './graph-utils'

export const GraphPage = () => {
  const modules = useModules()
  const [selectedNode, setSelectedNode] = useState<FTLNode | null>(null)
  const [selectedModuleName, setSelectedModuleName] = useState<string | null>(null)

  if (!modules.isSuccess) {
    return (
      <div className='flex justify-center items-center h-full'>
        <Loader />
      </div>
    )
  }

  const handleNodeTapped = (node: FTLNode | null, moduleName: string | null) => {
    setSelectedNode(node)
    setSelectedModuleName(moduleName)
  }

  return (
    <div className='flex h-full'>
      <ResizablePanels
        mainContent={<GraphPane onTapped={handleNodeTapped} />}
        rightPanelHeader={headerForNode(selectedNode, selectedModuleName)}
        rightPanelPanels={panelsForNode(selectedNode, selectedModuleName)}
        bottomPanelContent={<Timeline timeSettings={{ isTailing: true, isPaused: false }} filters={[]} />}
      />
    </div>
  )
}

const panelsForNode = (node: FTLNode | null, moduleName: string | null) => {
  if (node instanceof Module) {
    return modulePanels(node)
  }

  // If no module name is provided, we can't show the panels
  if (!moduleName) {
    return [] as ExpandablePanelProps[]
  }

  if (node instanceof Config) {
    return configPanels(moduleName, node, false)
  }
  if (node instanceof Secret) {
    return secretPanels(moduleName, node, false)
  }
  if (node instanceof Database) {
    return databasePanels(moduleName, node, false)
  }
  if (node instanceof Enum) {
    return enumPanels(moduleName, node, false)
  }
  if (node instanceof Data) {
    return dataPanels(moduleName, node, false)
  }
  if (node instanceof Topic) {
    return topicPanels(moduleName, node, false)
  }
  if (node instanceof Verb) {
    return verbPanels(moduleName, node, false)
  }
  return [] as ExpandablePanelProps[]
}
