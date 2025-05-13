import { CellsIcon, PackageIcon } from 'hugeicons-react'
import { Config, Data, Database, Enum, Module, Secret, Topic, Verb } from '../../protos/xyz/block/ftl/console/v1/console_pb'
import { RightPanelHeader } from '../modules/decls/RightPanelHeader'
import { declIcon, moduleUrlForRef } from '../modules/module.utils'
import type { FTLNode } from './graph-utils'

export const HeaderForNode = (node: FTLNode | null, ref?: string | null) => {
  if (!node || !ref) {
    return header({
      IconComponent: CellsIcon,
      content: <div className='text-sm font-medium truncate'>Root</div>,
    })
  }

  const urlHoverText = ref ? 'View in modules' : undefined

  if (node instanceof Module) {
    return <RightPanelHeader Icon={PackageIcon} title={node.name} url={moduleUrlForRef(ref, 'module')} urlHoverText={urlHoverText} />
  }
  if (node instanceof Verb) {
    if (!node.verb) return
    return <RightPanelHeader Icon={declIcon('verb', node.verb)} title={node.verb.name} url={moduleUrlForRef(ref, 'verb')} urlHoverText={urlHoverText} />
  }
  if (node instanceof Config) {
    if (!node.config) return
    return <RightPanelHeader Icon={declIcon('config', node.config)} title={node.config.name} url={moduleUrlForRef(ref, 'config')} urlHoverText={urlHoverText} />
  }
  if (node instanceof Secret) {
    if (!node.secret) return
    return <RightPanelHeader Icon={declIcon('secret', node.secret)} title={node.secret.name} url={moduleUrlForRef(ref, 'secret')} urlHoverText={urlHoverText} />
  }
  if (node instanceof Data) {
    if (!node.data) return
    return <RightPanelHeader Icon={declIcon('data', node.data)} title={node.data.name} />
  }
  if (node instanceof Database) {
    if (!node.database) return
    return (
      <RightPanelHeader
        Icon={declIcon('database', node.database)}
        title={node.database.name}
        url={moduleUrlForRef(ref, 'database')}
        urlHoverText={urlHoverText}
      />
    )
  }
  if (node instanceof Enum) {
    if (!node.enum) return
    return <RightPanelHeader Icon={declIcon('enum', node.enum)} title={node.enum.name} />
  }
  if (node instanceof Topic) {
    if (!node.topic) return
    return <RightPanelHeader Icon={declIcon('topic', node.topic)} title={node.topic.name} url={moduleUrlForRef(ref, 'topic')} urlHoverText={urlHoverText} />
  }
}

const header = ({ IconComponent, content }: { IconComponent: React.ElementType; content: React.ReactNode }) => {
  return (
    <div className='flex items-center gap-2 px-2 py-2'>
      <IconComponent className='h-5 w-5 text-indigo-600' />
      <div className='flex flex-col min-w-0'>{content}</div>
    </div>
  )
}
