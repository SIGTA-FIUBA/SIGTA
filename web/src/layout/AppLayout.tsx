import { Alert, AppShell, Container, Group, Text } from '@mantine/core'
import { useState } from 'react'
import { Outlet, useLoaderData, useNavigate } from 'react-router'
import { logout, type requireSession } from '../session'
import { UserMenu } from './UserMenu'

// AppLayout frames every page that needs a session: the name of the system
// and the account menu on top, the page below.
export function AppLayout() {
  const user = useLoaderData<typeof requireSession>()
  const navigate = useNavigate()
  const [logoutFailed, setLogoutFailed] = useState(false)

  async function handleLogout() {
    if (await logout()) {
      await navigate('/login', { replace: true })
    } else {
      setLogoutFailed(true)
    }
  }

  return (
    <AppShell header={{ height: 56 }} padding="md">
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Text fw={700} size="lg">
            SIGTA
          </Text>
          <UserMenu user={user} onLogout={() => void handleLogout()} />
        </Group>
      </AppShell.Header>
      <AppShell.Main>
        <Container size="md">
          {logoutFailed && (
            <Alert color="red" mb="md">
              No se pudo cerrar la sesión. Volvé a intentarlo.
            </Alert>
          )}
          <Outlet />
        </Container>
      </AppShell.Main>
    </AppShell>
  )
}
