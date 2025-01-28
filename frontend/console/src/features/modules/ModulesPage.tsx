import type React from 'react'
import { useMemo, useState } from 'react'
import { ResizableHorizontalPanels } from '../../shared/components/ResizableHorizontalPanels'
import { ModulesTree } from './ModulesTree'
import { useStreamModules } from './hooks/use-stream-modules'
import { getTreeWidthFromLS, moduleTreeFromStream, setTreeWidthInLS } from './module.utils'

export const ModulesPage = ({ body }: { body: React.ReactNode }) => {
  const modules = useStreamModules()
  const tree = useMemo(() => moduleTreeFromStream(modules?.data?.modules || []), [modules?.data])
  const [treeWidth, setTreeWidth] = useState(getTreeWidthFromLS())

  const setTreeWidthWithLS = (newWidth: number) => {
    const constrainedWidth = Math.max(200, newWidth)
    setTreeWidthInLS(constrainedWidth)
    setTreeWidth(constrainedWidth)
  }

  return (
    <div className='h-full'>
      <ResizableHorizontalPanels
        leftPanelContent={<ModulesTree modules={tree} />}
        rightPanelContent={body}
        leftPanelWidth={treeWidth}
        setLeftPanelWidth={setTreeWidthWithLS}
      />
    </div>
  )
}
