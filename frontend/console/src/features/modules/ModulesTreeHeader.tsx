import { ArrowShrink02Icon, ViewIcon, ViewOffSlashIcon } from 'hugeicons-react'
import { Button } from '../../shared/components/Button'
import { Multiselect } from '../../shared/components/Multiselect'
import type { MultiselectOpt } from '../../shared/components/Multiselect'
import { declTypeMultiselectOpts } from './schema/schema.utils'

interface ModuleTreeHeaderProps {
  selectedDeclTypes: MultiselectOpt[]
  showExported: boolean
  onDeclTypesChange: (opts: MultiselectOpt[]) => void
  onShowExportedChange: (val: boolean) => void
  onCollapseAll: () => void
}

export const ModuleTreeHeader = ({ selectedDeclTypes, showExported, onDeclTypesChange, onShowExportedChange, onCollapseAll }: ModuleTreeHeaderProps) => {
  return (
    <div className='border-b border-gray-120 dark:border-gray-700'>
      <div className='flex items-center gap-1 p-2 bg-white dark:bg-gray-800 shadow-sm'>
        <div className='flex-1 min-w-0 h-6'>
          <Multiselect allOpts={declTypeMultiselectOpts} selectedOpts={selectedDeclTypes} onChange={onDeclTypesChange} />
        </div>
        <div className='flex gap-1'>
          <Button
            id='show-exported'
            variant='secondary'
            size='xs'
            onClick={() => onShowExportedChange(!showExported)}
            title={showExported ? 'Show all (exported and unexported)' : 'Show only exported'}
          >
            {showExported ? <ViewOffSlashIcon className='size-4 text-red-400' /> : <ViewIcon className='size-4' />}
          </Button>
          <Button variant='secondary' size='xs' onClick={onCollapseAll} title='Collapse all modules'>
            <ArrowShrink02Icon className='size-4' />
          </Button>
        </div>
      </div>
    </div>
  )
}
