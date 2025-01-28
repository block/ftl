import type { Data } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import { ResizablePanels } from '../../../../shared/components/ResizablePanels'
import { declIcon } from '../../module.utils'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { dataPanels } from './DataRightPanels'

export const DataPanel = ({ data, moduleName, declName }: { data: Data; schema: string; moduleName: string; declName: string }) => {
  if (!data) {
    return
  }
  const decl = data.data
  if (!decl) {
    return
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <PanelHeader title='Data' declRef={`${moduleName}.${declName}`} exported={decl.export} comments={decl.comments} />
          </div>
        }
        rightPanelHeader={<RightPanelHeader Icon={declIcon('data', decl)} title={declName} />}
        rightPanelPanels={dataPanels(moduleName, data)}
        storageKeyPrefix='dataPanel'
      />
    </div>
  )
}
