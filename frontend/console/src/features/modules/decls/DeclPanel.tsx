import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import type { Config, Data, Database, Enum, Secret, Topic, TypeAlias } from '../../../protos/xyz/block/ftl/console/v1/console_pb'
import { useStreamModules } from '../hooks/use-stream-modules'
import { declFromModules } from '../module.utils'
import { ConfigPanel } from './config/ConfigPanel'
import { DataPanel } from './data/DataPanel'
import { DatabasePanel } from './database/DatabasePanel'
import { EnumPanel } from './enum/EnumPanel'
import { SecretPanel } from './secret/SecretPanel'
import { TopicPanel } from './topic/TopicPanel'
import { TypeAliasPanel } from './typealias/TypeAliasPanel'
import { VerbPage } from './verb/VerbPage'

export const DeclPanel = () => {
  const { moduleName, declCase, declName } = useParams()
  if (!moduleName || !declCase || !declName) {
    // Should be impossible, but validate anyway for type safety
    return
  }

  const modules = useStreamModules()
  const decl = useMemo(() => declFromModules(moduleName, declCase, declName, modules?.data?.modules), [modules?.data, moduleName, declCase, declName])
  if (!decl) {
    return
  }

  const nameProps = { moduleName, declName }
  const commonProps = { moduleName, declName, schema: decl.schema }
  switch (declCase) {
    case 'config':
      return <ConfigPanel {...commonProps} config={decl as Config} />
    case 'data':
      return <DataPanel {...commonProps} data={decl as Data} />
    case 'database':
      return <DatabasePanel {...commonProps} database={decl as Database} />
    case 'enum':
      return <EnumPanel {...commonProps} enumValue={decl as Enum} />
    case 'secret':
      return <SecretPanel {...commonProps} secret={decl as Secret} />
    case 'topic':
      return <TopicPanel {...commonProps} topic={decl as Topic} />
    case 'typealias':
      return <TypeAliasPanel {...commonProps} typealias={decl as TypeAlias} />
    case 'verb':
      return <VerbPage {...nameProps} />
  }
  return (
    <div className='flex-1 py-2 px-4'>
      <p>
        {declCase} declaration: {moduleName}.{declName}
      </p>
    </div>
  )
}
