import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from '@headlessui/react'
import { ArrowDown01Icon, CheckmarkSquare02Icon, MinusSignSquareIcon, SquareIcon } from 'hugeicons-react'
import { bgColor, borderColor } from '../utils'
import { textColor } from '../utils/style.utils'
import { Button } from './Button'

export interface MultiselectOpt {
  group?: string
  key: string
  displayName: string
}

export function sortMultiselectOpts(o: MultiselectOpt[]) {
  return o.sort((a: MultiselectOpt, b: MultiselectOpt) => (a.key < b.key ? -1 : 1))
}

const getSelectionText = (selectedOpts: MultiselectOpt[], allOpts: MultiselectOpt[]): string => {
  if (selectedOpts.length === 0) {
    return 'Select types...'
  }
  if (selectedOpts.length === allOpts.length) {
    return 'Filter types...'
  }
  return selectedOpts.map((o) => o.displayName).join(', ')
}

function getGroupsFromOpts(opts: MultiselectOpt[]): string[] {
  return [...new Set(opts.map((o) => o.group).filter((g) => !!g))] as string[]
}

const GroupIcon = ({ group, allOpts, selectedOpts }: { group: string; allOpts: MultiselectOpt[]; selectedOpts: MultiselectOpt[] }) => {
  const all = allOpts.filter((o) => o.group === group)
  const selected = selectedOpts.filter((o) => o.group === group)
  if (selected.length === 0) {
    return <SquareIcon className='size-5' />
  }
  if (all.length !== selected.length) {
    return <MinusSignSquareIcon className='size-5' />
  }
  return <CheckmarkSquare02Icon className='size-5' />
}

const optionClassName = (p: string) =>
  `cursor-pointer py-1 px-${p} group flex items-center gap-2 select-none text-sm text-gray-800 dark:text-gray-200 hover:bg-gray-200 hover:dark:bg-gray-700`

const Option = ({ o, p }: { o: MultiselectOpt; p: string }) => (
  <ListboxOption className={optionClassName(p)} value={o}>
    {({ selected }) => (
      <div className='flex items-center gap-2'>
        {selected ? <CheckmarkSquare02Icon className='size-5' /> : <SquareIcon className='size-5' />}
        {o.displayName}
      </div>
    )}
  </ListboxOption>
)

export const Multiselect = ({
  allOpts,
  selectedOpts,
  onChange,
}: { allOpts: MultiselectOpt[]; selectedOpts: MultiselectOpt[]; onChange: (types: MultiselectOpt[]) => void }) => {
  sortMultiselectOpts(selectedOpts)

  const groups = getGroupsFromOpts(allOpts)
  function toggleGroup(group: string) {
    const selected = selectedOpts.filter((o) => o.group === group)
    const xgroupSelectedOpts = selectedOpts.filter((o) => o.group !== group)
    if (selected.length === 0) {
      // Select all in group
      const allInGroup = allOpts.filter((o) => o.group === group)
      onChange([...xgroupSelectedOpts, ...allInGroup])
    } else {
      // Deselect all in group
      onChange(xgroupSelectedOpts)
    }
  }

  return (
    <div className='w-full'>
      <Listbox multiple value={selectedOpts} onChange={onChange}>
        <div className='relative w-full'>
          <ListboxButton
            className={`relative w-full cursor-pointer rounded-md ${bgColor} ${textColor} py-1 pl-2 pr-10 text-xs text-left shadow-sm ring-1 ring-inset ${borderColor} focus:outline-none focus:ring-2 focus:ring-indigo-600`}
          >
            <span className='block truncate'>{getSelectionText(selectedOpts, allOpts)}</span>
            <span className='pointer-events-none absolute inset-y-0 right-0 flex items-center pr-1'>
              <ArrowDown01Icon className='h-5 w-5 text-gray-400' aria-hidden='true' />
            </span>
          </ListboxButton>
        </div>
        <ListboxOptions
          anchor='bottom'
          transition
          className='w-[var(--button-width)] min-w-48 mt-1 pt-1 rounded-md border dark:border-white/5 bg-white dark:bg-gray-800 transition duration-100 ease-in truncate drop-shadow-lg z-20'
        >
          {allOpts
            .filter((o) => !o.group)
            .map((o) => (
              <Option key={o.key} o={o} p='2' />
            ))}
          {groups.map((group) => [
            <div key={group} onClick={() => toggleGroup(group)} className={optionClassName('2')}>
              <GroupIcon group={group} allOpts={allOpts} selectedOpts={selectedOpts} />
              {group}
            </div>,
            ...allOpts.filter((o) => o.group === group).map((o) => <Option key={o.key} o={o} p='6' />),
          ])}

          <div className='w-full text-center text-xs'>
            <div className='flex gap-2 p-2'>
              <Button variant='primary' size='xs' onClick={() => onChange(allOpts)} title='Select all' fullWidth>
                Select all
              </Button>
              <Button variant='primary' size='xs' onClick={() => onChange([])} title='Deselect all' fullWidth>
                Deselect all
              </Button>{' '}
            </div>
          </div>
        </ListboxOptions>
      </Listbox>
    </div>
  )
}
