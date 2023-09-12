import { PropsWithChildren, createContext, useContext, useEffect, useState } from 'react'
import { useClient } from '../hooks/use-client'
import { ConsoleService } from '../protos/xyz/block/ftl/v1/console/console_connect'
import { GetModulesResponse } from '../protos/xyz/block/ftl/v1/console/console_pb'
import { schemaContext } from './schema-provider'

export const modulesContext = createContext<GetModulesResponse>(new GetModulesResponse())

export const ModulesProvider = (props: PropsWithChildren) => {
  const schema = useContext(schemaContext)
  const client = useClient(ConsoleService)
  const [modules, setModules] = useState<GetModulesResponse>(new GetModulesResponse())

  useEffect(() => {
    const fetchModules = async () => {
      const modules = await client.getModules({})
      setModules(modules ?? [])

      return
    }

    fetchModules()
  }, [client, schema])

  return <modulesContext.Provider value={modules}>{props.children}</modulesContext.Provider>
}
