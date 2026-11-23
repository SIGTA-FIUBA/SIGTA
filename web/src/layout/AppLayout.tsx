import {
  Alert,
  Anchor,
  AppShell,
  Badge,
  Box,
  Burger,
  Button,
  Container,
  Image,
  Indicator,
  NavLink,
  Paper,
  Stack,
  Tooltip,
} from '@mantine/core'
import { useDisclosure, useMediaQuery } from '@mantine/hooks'
import { motion } from 'motion/react'
import { useState } from 'react'
import { Link, Outlet, useLoaderData, useLocation, useNavigate } from 'react-router'
import { useInboxProcedures } from '../api/procedures'
import { useEditableProcedureTypes } from '../api/procedureTypes'
import { logout, type optionalSession } from '../api/session'
import { Icon, type IconName } from '../components/Icon/Icon'
import { move } from '../theme/motion'
import { tokens } from '../theme/tokens'
import { AppHeader } from './AppHeader'
import { UserMenu } from './UserMenu'

type MenuLink = { to: string; label: string; icon: IconName; active: boolean; actionableCount?: number }

// CurrentMark is the mark of the current page in the menu, behind its item.
// There is one: going to another page, it travels from the item left to the
// one opened, so the menu shows the move and not only where the person is.
function CurrentMark() {
  return (
    <Paper
      variant="nav-current"
      aria-hidden
      renderRoot={(props) => <motion.div {...props} layoutId="nav-current" transition={move} />}
    />
  )
}

// Between breakpoint.sm and breakpoint.md the menu shows only its icons.
const compactQuery = `(min-width: ${tokens.breakpoints.sm}) and (max-width: calc(${tokens.breakpoints.md} - 0.1px))`

// AppLayout frames every page. With a session, a white side navigation runs
// from top to bottom: the SIGTA logo, the menu, and at the foot the page
// about SIGTA and the account. From breakpoint.md it is 240 wide; on a tablet
// it narrows to its icons; on a phone it folds behind a button in a white top
// bar. A visitor without a session, who can only read the catalog, gets the
// top bar with a way to log in and no menu.
export function AppLayout() {
  const user = useLoaderData<typeof optionalSession>()
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const [menuOpen, { toggle: toggleMenu, close: closeMenu }] = useDisclosure(false)
  const [logoutFailed, setLogoutFailed] = useState(false)
  const compact = useMediaQuery(compactQuery) ?? false
  const staff = user?.staff === true
  // The menu shows how many procedures need action; the inbox shares the
  // request, so it costs nothing more.
  const procedures = useInboxProcedures(staff)
  const actionableCount = procedures.data?.items.filter((item) => item.actionable).length
  // The procedure types are offered to whoever can edit or approve one; the
  // page shares the request.
  const editable = useEditableProcedureTypes({ enabled: staff, staleTime: 60_000 })
  const designs = editable.data !== undefined && (editable.data.can_create || editable.data.procedure_types.length > 0)

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
            actionableCount,
            active: pathname.startsWith('/inbox') || pathname.startsWith('/procedures'),
          },
        ]
      : []),
    { to: '/catalog', label: 'Catálogo de trámites', icon: 'catalog', active: pathname.startsWith('/catalog') },
    ...(designs
      ? [
          {
            to: '/procedure-types',
            label: 'Tipos de trámite',
            icon: 'procedureTypes' as const,
            active: pathname.startsWith('/procedure-types'),
          },
        ]
      : []),
    ...(user?.administrator === true
      ? [
          {
            to: '/administration',
            label: 'Administración',
            icon: 'administration' as const,
            active: pathname.startsWith('/administration'),
          },
        ]
      : []),
  ]

  function menuItem(link: MenuLink) {
    const count = link.actionableCount !== undefined && link.actionableCount > 0 ? link.actionableCount : undefined
    const label = count === undefined ? link.label : `${link.label}, ${count} con acción requerida`
    const common = {
      component: Link,
      to: link.to,
      active: link.active,
      'aria-current': link.active ? ('page' as const) : undefined,
      'aria-label': label,
      onClick: closeMenu,
    }
    if (compact) {
      return (
        <Box key={link.to} pos="relative">
          {link.active && <CurrentMark />}
          <Tooltip label={label} position="right">
            <NavLink
              {...common}
              variant="menu-icon"
              leftSection={
                <Indicator disabled={count === undefined} size={tokens.component.navItem.dot} color="fiuba.5">
                  <Icon name={link.icon} />
                </Indicator>
              }
            />
          </Tooltip>
        </Box>
      )
    }
    return (
      <Box key={link.to} pos="relative">
        {link.active && <CurrentMark />}
        <NavLink
          {...common}
          variant="menu"
          label={link.label}
          leftSection={<Icon name={link.icon} />}
          rightSection={
            count !== undefined && (
              // A long count would widen the badge past the menu; the label
              // keeps the exact number.
              <Badge variant="counter" aria-hidden>
                {count > 99 ? '99+' : count}
              </Badge>
            )
          }
        />
      </Box>
    )
  }

  const about: MenuLink = { to: '/about', label: 'Acerca de SIGTA', icon: 'info', active: pathname === '/about' }

  if (user === null) {
    return (
      <AppShell header={{ height: tokens.layout.topBar }} padding={{ base: 'md', sm: tokens.layout.pagePadding }}>
        <AppShell.Header>
          <AppHeader
            end={
              <Button component="a" href="/api/auth/login" variant="outline">
                Ingresar
              </Button>
            }
          />
        </AppShell.Header>
        <AppShell.Main>
          <Container px={0}>
            <Outlet context={user} />
          </Container>
        </AppShell.Main>
      </AppShell>
    )
  }

  return (
    <AppShell
      header={{ height: { base: tokens.layout.topBar, sm: 0 } }}
      navbar={{
        width: { sm: tokens.layout.navCompact, md: tokens.layout.nav },
        breakpoint: 'sm',
        collapsed: { mobile: !menuOpen },
      }}
      padding={{ base: 'md', sm: tokens.layout.pagePadding }}
    >
      <AppShell.Header hiddenFrom="sm">
        <AppHeader
          start={
            <Burger
              opened={menuOpen}
              onClick={toggleMenu}
              size="sm"
              aria-label={menuOpen ? 'Cerrar el menú' : 'Abrir el menú'}
            />
          }
        />
      </AppShell.Header>
      <AppShell.Navbar p={compact ? 'xs' : 'sm'} aria-label="Menú principal">
        <AppShell.Section visibleFrom="sm" px={compact ? 0 : 'xs'} pt="xs" pb="xl">
          <Anchor component={Link} to="/" display="flex" aria-label="SIGTA, ir al inicio">
            <Image
              src={compact ? '/brand/sigta-isotipo-color.svg' : '/brand/sigta-horizontal-color.svg'}
              alt=""
              h="var(--sigta-logo-height)"
              w="auto"
              mx={compact ? 'auto' : 0}
            />
          </Anchor>
        </AppShell.Section>
        <AppShell.Section grow>
          <Stack gap={0}>{links.map(menuItem)}</Stack>
        </AppShell.Section>
        <AppShell.Section>
          <Stack gap="xs">
            {menuItem(about)}
            <UserMenu user={user} compact={compact} onLogout={() => void handleLogout()} />
          </Stack>
        </AppShell.Section>
      </AppShell.Navbar>
      <AppShell.Main>
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
