import { Notification } from '../layout/Notification'
import { InfoProvider } from './info-provider'
import { NotificationsProvider } from './notifications-provider'
import { ReactQueryProvider } from './react-query-provider'
import { RoutingProvider } from './routing-provider'
import { UserPreferencesProvider } from './user-preferences-provider'

export const AppProvider = () => {
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
