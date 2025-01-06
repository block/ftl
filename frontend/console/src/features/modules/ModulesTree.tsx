import { ArrowRight01Icon, ArrowShrink02Icon, CircleArrowRight02Icon, ViewIcon, ViewOffSlashIcon } from 'hugeicons-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { Button } from '../../components/Button'
import { Multiselect, sortMultiselectOpts } from '../../components/Multiselect'
import type { MultiselectOpt } from '../../components/Multiselect'
import { classNames } from '../../utils'
import type { DeclInfo, ModuleTreeItem } from './module.utils'
import {
  addModuleToLocalStorageIfMissing,
  collapseAllModulesInLocalStorage,
  declIcon,
  declSumTypeIsExported,
  declTypeName,
  declUrlFromInfo,
  getExpandedDeclTypesFromLocalStorage,
  getHideUnexportedFromLocalStorage,
  hasHideUnexportedInLocalStorage,
  listExpandedModulesFromLocalStorage,
  setExpandedDeclTypesInLocalStorage,
  setHideUnexportedFromLocalStorage,
  toggleModuleExpansionInLocalStorage,
} from './module.utils'
import { declTypeMultiselectOpts } from './schema/schema.utils'

const ModuleSection = ({
  module,
  isExpanded,
  toggleExpansion,
  selectedDeclTypes,
  hideUnexported,
  expandedDeclTypes,
  toggleDeclType,
}: {
  module: ModuleTreeItem
  isExpanded: boolean
  toggleExpansion: (m: string) => void
  selectedDeclTypes: MultiselectOpt[]
  hideUnexported: boolean
  expandedDeclTypes: string[]
  toggleDeclType: (moduleName: string, declType: string) => void
}) => {
  const { moduleName, declName } = useParams()
  const isSelected = useMemo(() => moduleName === module.name, [moduleName, module.name])
  const moduleRef = useRef<HTMLDivElement>(null)

  // Scroll to the selected module on page load
  useEffect(() => {
    if (isSelected && !declName && moduleRef.current) {
      const { top } = moduleRef.current.getBoundingClientRect()
      const { innerHeight } = window
      if (top < 64 || top > innerHeight) {
        moduleRef.current.scrollIntoView()
      }
    }
  }, [moduleName]) // moduleName is the selected module; module.name is the one being rendered

  const filteredDecls = useMemo(
    () =>
      module.decls
        .filter((d) => !!selectedDeclTypes.find((o) => o.key === declTypeName(d.declType, d.value)))
        .filter((d) => !hideUnexported || (isSelected && declName === d.value.name) || declSumTypeIsExported(d.value)),
    [module.decls, selectedDeclTypes, hideUnexported, isSelected, declName],
  )

  // Group declarations by their type
  const groupedDecls = useMemo(() => {
    const groups: Record<string, DeclInfo[]> = {}
    for (const decl of filteredDecls) {
      const type = declTypeName(decl.declType, decl.value)
      if (!groups[type]) {
        groups[type] = []
      }
      groups[type].push(decl)
    }
    return groups
  }, [filteredDecls])

  const [expandedGroups, setExpandedGroups] = useState<string[]>(() => {
    if (isSelected && declName) {
      const selectedDeclType = module.decls.find((d) => d.value.name === declName)
      if (selectedDeclType) {
        return [declTypeName(selectedDeclType.declType, selectedDeclType.value)]
      }
    }
    return []
  })

  useEffect(() => {
    if (isSelected && declName) {
      const selectedDeclType = module.decls.find((d) => d.value.name === declName)
      if (selectedDeclType) {
        const typeName = declTypeName(selectedDeclType.declType, selectedDeclType.value)
        if (!expandedGroups.includes(typeName)) {
          setExpandedGroups((prev) => [...prev, typeName])
        }
      }
    }
  }, [isSelected, declName, module.decls, expandedGroups])

  return (
    <li key={module.name} id={`module-tree-module-${module.name}`} className='mb-2'>
      <div
        ref={moduleRef}
        id={`module-${module.name}-tree-group`}
        className={classNames(
          isSelected ? 'bg-gray-100 dark:bg-gray-700 hover:bg-gray-300 hover:dark:bg-gray-600' : 'hover:bg-gray-200 hover:dark:bg-gray-700',
          'group flex w-full items-center gap-x-2 text-left text-sm font-medium cursor-pointer leading-6 rounded-md px-2',
        )}
        onClick={() => toggleExpansion(module.name)}
      >
        <ArrowRight01Icon aria-hidden='true' className={`h-4 w-4 shrink-0 ${isExpanded ? 'rotate-90 text-gray-500' : ''}`} />
        {module.name}
        <Link to={`/modules/${module.name}`} onClick={(e) => e.stopPropagation()}>
          <CircleArrowRight02Icon id={`module-${module.name}-view-icon`} className='size-4 shrink-0 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600' />
        </Link>
      </div>
      {isExpanded && (
        <ul className='pl-4'>
          {Object.entries(groupedDecls).map(([groupName, decls]) => {
            const declTypeKey = `${module.name}:${groupName}`
            const isGroupExpanded = expandedDeclTypes.includes(declTypeKey)
            const DeclTypeIcon = declIcon(decls[0].declType, decls[0].value)

            return (
              <li key={groupName} className='my-1'>
                <div
                  data-test-id='module-tree-group'
                  data-group-type={groupName}
                  className={classNames(
                    'group flex w-full items-center gap-x-2 text-left text-sm font-medium cursor-pointer leading-6 rounded-md hover:bg-gray-200 dark:hover:bg-gray-700',
                  )}
                  onClick={() => toggleDeclType(module.name, groupName)}
                >
                  <ArrowRight01Icon aria-hidden='true' className={`ml-1 h-4 w-4 shrink-0 ${isGroupExpanded ? 'rotate-90 text-gray-500' : ''}`} />
                  <span title={groupName}>
                    <DeclTypeIcon aria-hidden='true' className='size-4 shrink-0' />
                  </span>
                  {groupName}
                  <span className='text-xs text-gray-500'>({decls.length})</span>
                </div>
                {isGroupExpanded && (
                  <ul>
                    {decls.map((d, i) => {
                      const DeclIcon = declIcon(d.declType, d.value)
                      return (
                        <li key={i} className='my-1'>
                          <Link id={`decl-${d.value.name}`} to={declUrlFromInfo(module.name, d)}>
                            <div
                              className={classNames(
                                isSelected && declName === d.value.name
                                  ? 'bg-gray-100 dark:bg-gray-700 hover:bg-gray-300 hover:dark:bg-gray-600'
                                  : 'hover:bg-gray-200 hover:dark:bg-gray-700',
                                declSumTypeIsExported(d.value) ? '' : 'text-gray-400 dark:text-gray-500',
                                'group flex items-center gap-x-2 pl-7 pr-2 text-sm font-light leading-6 w-full cursor-pointer scroll-mt-10 rounded-md',
                              )}
                            >
                              <span title={d.value.name}>
                                <DeclIcon aria-hidden='true' className='size-4 shrink-0' />
                              </span>
                              {d.value.name}
                            </div>
                          </Link>
                        </li>
                      )
                    })}
                  </ul>
                )}
              </li>
            )
          })}
        </ul>
      )}
    </li>
  )
}

const declTypesSearchParamKey = 'dt'

export const ModulesTree = ({ modules }: { modules: ModuleTreeItem[] }) => {
  const { moduleName, declName } = useParams()

  const [searchParams, setSearchParams] = useSearchParams()
  const declTypeKeysFromUrl = searchParams.getAll(declTypesSearchParamKey)
  const declTypesFromUrl = declTypeMultiselectOpts.filter((o) => declTypeKeysFromUrl.includes(o.key))
  const [selectedDeclTypes, setSelectedDeclTypes] = useState(declTypesFromUrl.length === 0 ? declTypeMultiselectOpts : declTypesFromUrl)

  const initialExpanded = listExpandedModulesFromLocalStorage()
  const [expandedModules, setExpandedModules] = useState(initialExpanded)

  useEffect(() => {
    if (moduleName && declName) {
      addModuleToLocalStorageIfMissing(moduleName)
    }
    setExpandedModules(listExpandedModulesFromLocalStorage())
  }, [moduleName, declName])

  const [hideUnexported, setHideUnexported] = useState(hasHideUnexportedInLocalStorage() ? getHideUnexportedFromLocalStorage() : true)

  function msOnChange(opts: MultiselectOpt[]) {
    const params = new URLSearchParams()
    if (opts.length !== declTypeMultiselectOpts.length) {
      for (const o of sortMultiselectOpts(opts)) {
        params.append(declTypesSearchParamKey, o.key)
      }
    }
    setSearchParams(params)
    setSelectedDeclTypes(opts)
  }

  function toggle(toggledModule: string) {
    toggleModuleExpansionInLocalStorage(toggledModule)
    setExpandedModules(listExpandedModulesFromLocalStorage())
  }

  function collapseAll() {
    // Collapse all modules
    collapseAllModulesInLocalStorage()
    if (moduleName && declName) {
      addModuleToLocalStorageIfMissing(moduleName)
    }
    setExpandedModules(listExpandedModulesFromLocalStorage())

    // Collapse all decl types
    setExpandedDeclTypes([])
    setExpandedDeclTypesInLocalStorage([])
  }

  function setHideUnexportedState(val: boolean) {
    setHideUnexportedFromLocalStorage(val)
    setHideUnexported(val)
  }

  const [expandedDeclTypes, setExpandedDeclTypes] = useState<string[]>(() => {
    return getExpandedDeclTypesFromLocalStorage()
  })

  function toggleDeclType(moduleName: string, declType: string) {
    const key = `${moduleName}:${declType}`
    const newExpanded = expandedDeclTypes.includes(key) ? expandedDeclTypes.filter((t) => t !== key) : [...expandedDeclTypes, key]
    setExpandedDeclTypes(newExpanded)
    setExpandedDeclTypesInLocalStorage(newExpanded)
  }

  modules.sort((m1, m2) => Number(m1.isBuiltin) - Number(m2.isBuiltin))

  return (
    <div className='flex flex-col h-full border-r border-gray-300 dark:border-gray-700'>
      <div className='border-b border-gray-120 dark:border-gray-700'>
        <div className='flex items-center gap-1 p-2 bg-white dark:bg-gray-800 shadow-sm'>
          <div className='flex-1 min-w-0 h-6'>
            <Multiselect allOpts={declTypeMultiselectOpts} selectedOpts={selectedDeclTypes} onChange={msOnChange} />
          </div>
          <div className='flex gap-1'>
            <Button id='hide-exported' variant='secondary' size='xs' onClick={() => setHideUnexportedState(!hideUnexported)} title='Show/hide unexported'>
              {hideUnexported ? <ViewOffSlashIcon className='size-4' /> : <ViewIcon className='size-4' />}
            </Button>
            <Button variant='secondary' size='xs' onClick={collapseAll} title='Collapse all modules'>
              <ArrowShrink02Icon className='size-4' />
            </Button>
          </div>
        </div>
      </div>
      <nav className='overflow-y-auto flex-1'>
        <ul id='module-tree-content' className='p-2'>
          {modules.map((m) => (
            <ModuleSection
              key={m.name}
              module={m}
              isExpanded={expandedModules.includes(m.name)}
              toggleExpansion={toggle}
              selectedDeclTypes={selectedDeclTypes}
              hideUnexported={hideUnexported}
              expandedDeclTypes={expandedDeclTypes}
              toggleDeclType={toggleDeclType}
            />
          ))}
        </ul>
      </nav>
    </div>
  )
}
