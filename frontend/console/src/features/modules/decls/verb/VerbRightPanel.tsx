import type { Verb } from '../../../../protos/xyz/block/ftl/console/v1/console_pb'
import type { ExpandablePanelProps } from '../../../../shared/components/ExpandablePanel'
import { RightPanelAttribute } from '../../../../shared/components/RightPanelAttribute'
import { DeclDefaultPanels } from '../DeclDefaultPanels'
import { httpRequestPath, ingress, isHttpIngress } from './verb.utils'

export const verbPanels = (moduleName: string, verb?: Verb) => {
  const panels = [] as ExpandablePanelProps[]

  if (isHttpIngress(verb)) {
    const http = ingress(verb)
    const path = httpRequestPath(verb)
    panels.push({
      title: 'HTTP Ingress',
      expanded: true,
      children: (
        <>
          <RightPanelAttribute name='Method' value={http.method} />
          <RightPanelAttribute name='Path' value={path} />
        </>
      ),
    })
  }

  panels.push(...DeclDefaultPanels(moduleName, verb?.schema, verb?.edges))

  return panels
}
