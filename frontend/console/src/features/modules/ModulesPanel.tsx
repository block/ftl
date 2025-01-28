import type { Module } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { AttributeBadge } from '../../shared/components/AttributeBadge'
import { List } from '../../shared/components/List'
import { classNames } from '../../shared/utils'
import { deploymentTextColor } from '../deployments/deployment.utils'
import { useModules } from './hooks/use-modules'

export const ModulesPanel = () => {
  const modules = useModules()

  const moduleHref = (module: Module) => `/modules/${module.name}`

  return (
    <div className='p-2'>
      <List
        items={modules.data?.modules ?? []}
        href={moduleHref}
        renderItem={(module) => (
          <div className='flex w-full' data-module-row={module.name}>
            <div className='flex gap-x-4 items-center w-1/2'>
              <div className='whitespace-nowrap'>
                <div className='flex gap-x-2 items-center'>
                  <p>{module.name}</p>
                </div>

                <p className={classNames(deploymentTextColor(module.deploymentKey), 'text-sm leading-6')}>{module.deploymentKey}</p>
              </div>
            </div>
            <div className='flex gap-x-4 items-center w-1/2 justify-end'>
              <div className='flex flex-wrap gap-2'>
                <AttributeBadge name='language' value={module.language} />
              </div>
            </div>
          </div>
        )}
      />
    </div>
  )
}
