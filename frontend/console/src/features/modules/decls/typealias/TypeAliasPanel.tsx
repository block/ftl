import type { TypeAlias } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import { Visibility } from '../../../../protos/xyz/block/ftl/schema/v1/schema_pb'
import { ResizablePanels } from '../../../../shared/components/ResizablePanels'
import { declIcon } from '../../module.utils'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { typeAliasPanels } from './TypeAliasRightPanels'

export const TypeAliasPanel = ({ typealias, moduleName, declName }: { typealias: TypeAlias; moduleName: string; declName: string }) => {
  if (!typealias) {
    return
  }

  const decl = typealias.typealias
  if (!decl) {
    return
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <PanelHeader
              title='TypeAlias'
              declRef={`${moduleName}.${declName}`}
              exported={decl.visibility === Visibility.SCOPE_MODULE || decl.visibility === Visibility.SCOPE_REALM}
              comments={decl.comments}
            />
          </div>
        }
        rightPanelHeader={<RightPanelHeader Icon={declIcon('typealias', decl)} title={declName} />}
        rightPanelPanels={typeAliasPanels(moduleName, typealias)}
        storageKeyPrefix='typeAliasPanel'
      />
    </div>
  )
}
