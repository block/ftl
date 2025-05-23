import { Navigate, Route, RouterProvider, createBrowserRouter, createRoutesFromElements } from 'react-router-dom'
import { GraphPage } from '../../features/graph/GraphPage'
import { InfrastructurePage } from '../../features/infrastructure/InfrastructurePage'
import { ModulePanel } from '../../features/modules/ModulePanel'
import { ModulesPage } from '../../features/modules/ModulesPage'
import { ModulesPanel } from '../../features/modules/ModulesPanel'
import { DeclPanel } from '../../features/modules/decls/DeclPanel'
import { TimelinePage } from '../../features/timeline/TimelinePage'
import { TracesPage } from '../../features/traces/TracesPage'
import { Layout } from '../layout/Layout'
import { NotFoundPage } from '../layout/NotFoundPage'

const router = createBrowserRouter(
  createRoutesFromElements(
    <>
      <Route path='/' element={<Layout />}>
        <Route index element={<Navigate to='events' replace />} />
        <Route path='events' element={<TimelinePage />} />
        <Route path='modules' element={<ModulesPage body={<ModulesPanel />} />} />
        <Route path='modules/:moduleName' element={<ModulesPage body={<ModulePanel />} />} />
        <Route path='modules/:moduleName/:declCase/:declName' element={<ModulesPage body={<DeclPanel />} />} />
        <Route path='graph' element={<GraphPage />} />
        <Route path='graph/:moduleName' element={<GraphPage />} />
        <Route path='graph/:moduleName/:declName' element={<GraphPage />} />
        <Route path='infrastructure' element={<Navigate to='deployments' replace />} />
        <Route path='infrastructure/*' element={<InfrastructurePage />} />
        <Route path='traces/:requestKey' element={<TracesPage />} />
      </Route>

      <Route path='*' element={<NotFoundPage />} />
    </>,
  ),
)

export const RoutingProvider = () => {
  return <RouterProvider router={router} />
}
