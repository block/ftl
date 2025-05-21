import { ArrowRight01Icon } from 'hugeicons-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { HoverPopup } from '../../shared/components/HoverPopup'
import type { MultiselectOpt } from '../../shared/components/Multiselect'
import { useInfo } from '../../shared/providers/info-provider'
import { classNames } from '../../shared/utils'
import { BuildStatusIndicator } from '../engine/BuildStatusIndicator'
import { getEventText } from '../engine/engine.utils'
import { useEngineStatus } from '../engine/hooks/use-engine-status'
import { ModuleTreeHeader } from './ModulesTreeHeader'
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
  sortModules,
  toggleModuleExpansionInLocalStorage,
} from './module.utils'
import { declTypeMultiselectOpts } from './schema/schema.utils'

const ModuleSection = ({
  module,
  isExpanded,
  toggleExpansion,
  selectedDeclTypes,
  showExported,
  expandedDeclTypes,
  toggleDeclType,
}: {
  module: ModuleTreeItem
  isExpanded: boolean
  toggleExpansion: (m: string) => void
  selectedDeclTypes: MultiselectOpt[]
  showExported: boolean
  expandedDeclTypes: string[]
  toggleDeclType: (moduleName: string, declType: string) => void
}) => {
  const { moduleName, declName } = useParams()
  const isSelected = useMemo(() => moduleName === module.name, [moduleName, module.name])
  const moduleRef = useRef<HTMLDivElement>(null)
  const { isLocalDev } = useInfo()

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
        .filter((d) => !showExported || (isSelected && declName === d.value.name) || declSumTypeIsExported(d.value)),
    [module.decls, selectedDeclTypes, showExported, isSelected, declName],
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

  const { modules } = useEngineStatus()
  const moduleEvent = modules[module.name]

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
        <ArrowRight01Icon aria-hidden='true' className={`size-4 shrink-0 ${isExpanded ? 'rotate-90 text-gray-500' : ''}`} />
        <HoverPopup popupContent={getEventText(moduleEvent)}>
          <BuildStatusIndicator eventCase={isLocalDev ? moduleEvent?.event?.case : 'moduleDeploySuccess'} />
        </HoverPopup>
        {module.name}
        <Link
          to={`/modules/${module.name}`}
          onClick={(e) => e.stopPropagation()}
          className='opacity-0 group-hover:opacity-100 text-xs text-indigo-500 dark:text-indigo-300 hover:text-indigo-400 dark:hover:text-indigo-200 ml-auto'
        >
          view
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

export const ModulesTree = ({ modules }: { modules: ModuleTreeItem[] }) => {
  const { moduleName, declName } = useParams()

  const declTypeKeysFromUrl: string[] = [] // No longer using URL params
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

  // Use localStorage for showExported state
  const [showExported, setShowExported] = useState(hasHideUnexportedInLocalStorage() ? getHideUnexportedFromLocalStorage() : false)

  const msOnChange = (opts: MultiselectOpt[]) => {
    setSelectedDeclTypes(opts)
  }

  const toggle = (toggledModule: string) => {
    toggleModuleExpansionInLocalStorage(toggledModule)
    setExpandedModules(listExpandedModulesFromLocalStorage())
  }

  const collapseAll = () => {
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

  const setShowExportedState = (val: boolean) => {
    setHideUnexportedFromLocalStorage(val)
    setShowExported(val)
  }

  const [expandedDeclTypes, setExpandedDeclTypes] = useState<string[]>(() => {
    return getExpandedDeclTypesFromLocalStorage()
  })

  const toggleDeclType = (moduleName: string, declType: string) => {
    const key = `${moduleName}:${declType}`
    const newExpanded = expandedDeclTypes.includes(key) ? expandedDeclTypes.filter((t) => t !== key) : [...expandedDeclTypes, key]
    setExpandedDeclTypes(newExpanded)
    setExpandedDeclTypesInLocalStorage(newExpanded)
  }

  const sortedModules = useMemo(() => {
    return sortModules(modules)
  }, [modules])

  return (
    <div className='flex flex-col h-full border-r border-gray-300 dark:border-gray-700'>
      <ModuleTreeHeader
        selectedDeclTypes={selectedDeclTypes}
        showExported={showExported}
        onDeclTypesChange={msOnChange}
        onShowExportedChange={setShowExportedState}
        onCollapseAll={collapseAll}
      />
      <nav className='overflow-y-auto flex-1'>
        <ul id='module-tree-content' className='p-2'>
          {sortedModules.map((m) => (
            <ModuleSection
              key={m.name}
              module={m}
              isExpanded={expandedModules.includes(m.name)}
              toggleExpansion={toggle}
              selectedDeclTypes={selectedDeclTypes}
              showExported={showExported}
              expandedDeclTypes={expandedDeclTypes}
              toggleDeclType={toggleDeclType}
            />
          ))}
        </ul>
      </nav>
    </div>
  )
}
