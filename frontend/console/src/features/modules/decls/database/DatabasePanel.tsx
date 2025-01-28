import { ResizablePanels } from '../../../../components/ResizablePanels'
import type { Database } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import { declIcon } from '../../module.utils'
import { PanelHeader } from '../PanelHeader'
import { RightPanelHeader } from '../RightPanelHeader'
import { databasePanels } from './DatabaseRightPanels'

export const DatabasePanel = ({ database, moduleName, declName }: { database: Database; schema: string; moduleName: string; declName: string }) => {
  const decl = database.database
  if (!decl) {
    return
  }

  return (
    <div className='h-full'>
      <ResizablePanels
        mainContent={
          <div className='p-4'>
            <PanelHeader title='Database' declRef={`${moduleName}.${declName}`} exported={false} comments={decl.comments} />
          </div>
        }
        rightPanelHeader={<RightPanelHeader Icon={declIcon('database', decl)} title={declName} />}
        rightPanelPanels={databasePanels(moduleName, database)}
        storageKeyPrefix='databasePanel'
      />
    </div>
  )
}
