import { ResizablePanels } from '../../../../components/ResizablePanels'
import type { Enum } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import { declIcon } from '../../module.utils'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { enumPanels } from './EnumRightPanels'

export const EnumPanel = ({ enumValue, moduleName, declName }: { enumValue: Enum; moduleName: string; declName: string }) => {
  if (!enumValue) {
    return
  }
  const decl = enumValue.enum
  if (!decl) {
    return
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <PanelHeader title='Enum' declRef={`${moduleName}.${declName}`} exported={false} comments={decl.comments} />
          </div>
        }
        rightPanelHeader={<RightPanelHeader Icon={declIcon('enum', decl)} title={declName} />}
        rightPanelPanels={enumPanels(moduleName, enumValue)}
        storageKeyPrefix='enumPanel'
      />
    </div>
  )
}
