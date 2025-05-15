import { useEffect } from 'react'
import { Notification } from '../layout/Notification'
import { InfoProvider } from './info-provider'
import { NotificationsProvider } from './notifications-provider'
import { ReactQueryProvider } from './react-query-provider'
import { RoutingProvider } from './routing-provider'
import { UserPreferencesProvider } from './user-preferences-provider'

export const AppProvider = () => {
  // Prevent Ctrl+S from saving the page
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 's') {
        e.preventDefault()
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  return (
    <ReactQueryProvider>
      <InfoProvider>
        <UserPreferencesProvider>
          <NotificationsProvider>
            <RoutingProvider />
            <Notification />
          </NotificationsProvider>
        </UserPreferencesProvider>
      </InfoProvider>
    </ReactQueryProvider>
  )
}
