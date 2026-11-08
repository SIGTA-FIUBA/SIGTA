import {
  Alert,
  AppShell,
  Avatar,
  Badge,
  Burger,
  Button,
  Container,
  Group,
  NavLink,
  Paper,
  Stack,
  Text,
} from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'
import { useState } from 'react'
import { Link, Outlet, useLoaderData, useLocation, useNavigate } from 'react-router'
import { useProcedures } from '../api/procedures'
import { MenuIcon, type MenuIconName } from '../components/MenuIcon/MenuIcon'
import { logout, type optionalSession } from '../session'
import { AppHeader } from './AppHeader'
import { UserMenu } from './UserMenu'

const profileNames: Record<string, string> = {
  applicant: 'Titular',
  staff_operative: 'Perfil operativo',
  staff_approver: 'Perfil de aprobación',
}

type MenuLink = { to: string; label: string; icon: MenuIconName; active: boolean; pending?: number }

// AppLayout frames every page: the faculty's logo and the account menu on
// top, the menu on the left in the faculty's light blue, and the page beside
// it. On a phone the menu folds behind a button. A visitor without a session,
// who can only read the catalog, gets no menu and a way to log in.
export function AppLayout() {
  const user = useLoaderData<typeof optionalSession>()
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const [menuOpen, { toggle: toggleMenu, close: closeMenu }] = useDisclosure(false)
  const [logoutFailed, setLogoutFailed] = useState(false)
  const staff = user !== null && user.role !== 'applicant'
  // The menu shows how many procedures need action; the inbox shares the
  // request, so it costs nothing more.
  const procedures = useProcedures(staff)
  const pending = procedures.data?.items.filter((item) => item.actionable).length

  async function handleLogout() {
    if (await logout()) {
      await navigate('/login', { replace: true })
    } else {
      setLogoutFailed(true)
    }
  }

  const links: MenuLink[] = [
    { to: '/', label: 'Inicio', icon: 'home', active: pathname === '/' },
    ...(staff
      ? [
          {
            to: '/inbox',
            label: 'Bandeja de trámites',
            icon: 'inbox' as const,
            pending,
            active: pathname.startsWith('/inbox') || pathname.startsWith('/procedures'),
          },
        ]
      : []),
    { to: '/catalog', label: 'Catálogo de trámites', icon: 'catalog', active: pathname.startsWith('/catalog') },
  ]

  return (
    <AppShell
      header={{ height: 64 }}
      navbar={user === null ? undefined : { width: 240, breakpoint: 'sm', collapsed: { mobile: !menuOpen } }}
      padding="lg"
    >
      <AppShell.Header>
        <AppHeader
          start={
            user !== null && (
              <Burger
                opened={menuOpen}
                onClick={toggleMenu}
                hiddenFrom="sm"
                size="sm"
                aria-label={menuOpen ? 'Cerrar el menú' : 'Abrir el menú'}
              />
            )
          }
          end={
            user !== null ? (
              <UserMenu user={user} onLogout={() => void handleLogout()} />
            ) : (
              <>
                <Button component={Link} to="/login" variant="subtle">
                  Volver al inicio
                </Button>
                <Button component={Link} to="/login">
                  Iniciar sesión
                </Button>
              </>
            )
          }
        />
      </AppShell.Header>
      {user !== null && (
        <AppShell.Navbar p="sm" bg="fiuba.5" aria-label="Menú principal">
          <AppShell.Section grow>
            <Stack gap="xs">
              {links.map((link) => (
                <NavLink
                  key={link.to}
                  component={Link}
                  to={link.to}
                  label={link.label}
                  leftSection={<MenuIcon name={link.icon} />}
                  rightSection={
                    link.pending !== undefined && link.pending > 0 ? (
                      <Badge
                        variant={link.active ? 'white' : 'filled'}
                        size="sm"
                        circle
                        aria-label={`${link.pending} con acción requerida`}
                      >
                        {link.pending}
                      </Badge>
                    ) : undefined
                  }
                  active={link.active}
                  variant="menu"
                  fw={600}
                  onClick={closeMenu}
                />
              ))}
            </Stack>
          </AppShell.Section>
          <AppShell.Section>
            <Paper p="sm" radius="md" shadow="xs">
              <Group gap="sm" wrap="nowrap">
                <Avatar name={user.name} color="fiuba" aria-hidden />
                <Stack gap={0} miw={0}>
                  <Text size="sm" fw={600} truncate="end">
                    {user.name}
                  </Text>
                  <Text size="xs" c="dimmed">
                    {profileNames[user.role]}
                  </Text>
                </Stack>
              </Group>
            </Paper>
          </AppShell.Section>
        </AppShell.Navbar>
      )}
      <AppShell.Main bg="gray.0">
        <Container size="xl" px={0}>
          {logoutFailed && (
            <Alert color="red" mb="md">
              No se pudo cerrar la sesión. Volvé a intentarlo.
            </Alert>
          )}
          <Outlet context={user} />
        </Container>
      </AppShell.Main>
    </AppShell>
  )
}
