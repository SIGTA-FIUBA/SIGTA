import { Alert, AppShell, Badge, Burger, Button, Container, Indicator, NavLink, Stack, Tooltip } from '@mantine/core'
import { useDisclosure, useMediaQuery } from '@mantine/hooks'
import { useState } from 'react'
import { Link, Outlet, useLoaderData, useLocation, useNavigate } from 'react-router'
import { useProcedures } from '../api/procedures'
import { Icon, type IconName } from '../components/Icon/Icon'
import { logout, type optionalSession } from '../session'
import { tokens } from '../theme/tokens'
import { AppHeader } from './AppHeader'
import { UserMenu } from './UserMenu'

type MenuLink = { to: string; label: string; icon: IconName; active: boolean; pending?: number }

// Between breakpoint.sm and breakpoint.md the menu shows only its icons.
const compactQuery = `(min-width: ${tokens.breakpoints.sm}) and (max-width: calc(${tokens.breakpoints.md} - 0.1px))`

// AppLayout frames every page: the SIGTA signature and the account menu on
// top, the menu on the left and the page beside it. From breakpoint.md the
// menu is 240 wide; on a tablet it narrows to its icons; on a phone it folds
// behind a button. A visitor without a session, who can only read the
// catalog, gets no menu and a way to log in.
export function AppLayout() {
  const user = useLoaderData<typeof optionalSession>()
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const [menuOpen, { toggle: toggleMenu, close: closeMenu }] = useDisclosure(false)
  const [logoutFailed, setLogoutFailed] = useState(false)
  const compact = useMediaQuery(compactQuery) ?? false
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

  function menuItem(link: MenuLink) {
    const count = link.pending !== undefined && link.pending > 0 ? link.pending : undefined
    const common = {
      component: Link,
      to: link.to,
      active: link.active,
      'aria-current': link.active ? ('page' as const) : undefined,
      onClick: closeMenu,
    }
    if (compact) {
      const label = count === undefined ? link.label : `${link.label}, ${count} con acción requerida`
      return (
        <Tooltip key={link.to} label={label} position="right">
          <NavLink
            {...common}
            variant="menu-icon"
            aria-label={label}
            leftSection={
              <Indicator disabled={count === undefined} size={tokens.component.navItem.dot}>
                <Icon name={link.icon} />
              </Indicator>
            }
          />
        </Tooltip>
      )
    }
    return (
      <NavLink
        key={link.to}
        {...common}
        variant="menu"
        label={link.label}
        leftSection={<Icon name={link.icon} />}
        rightSection={
          count !== undefined && (
            <Badge variant="filled" size="sm" circle aria-label={`${count} con acción requerida`}>
              {count}
            </Badge>
          )
        }
      />
    )
  }

  return (
    <AppShell
      header={{ height: tokens.layout.topBar }}
      navbar={
        user === null
          ? undefined
          : {
              width: { sm: tokens.layout.navCompact, md: tokens.layout.nav },
              breakpoint: 'sm',
              collapsed: { mobile: !menuOpen },
            }
      }
      padding={{ base: 'md', sm: 'xl' }}
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
        <AppShell.Navbar p="sm" aria-label="Menú principal">
          <Stack gap="xxs">{links.map(menuItem)}</Stack>
        </AppShell.Navbar>
      )}
      <AppShell.Main bg="gray.0">
        <Container px={0}>
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
