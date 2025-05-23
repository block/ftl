import { Outlet } from 'react-router-dom'
import { useInfo } from '../providers/info-provider'
import { Navigation } from './navigation/Navigation'

export const Layout = () => {
  const info = useInfo()

  return (
    <div className='min-w-[700px] max-w-full max-h-full h-full flex flex-col dark:bg-gray-800 bg-white text-gray-700 dark:text-gray-200'>
      <Navigation version={info.infoData?.version} />
      <main className='flex-1' style={{ height: 'calc(100vh - 64px)' }}>
        <Outlet />
      </main>
    </div>
  )
}
