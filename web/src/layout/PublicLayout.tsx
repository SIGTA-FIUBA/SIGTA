import { Anchor, Button, Group, Image, Paper, Stack } from '@mantine/core'
import { AnimatePresence, motion } from 'motion/react'
import { useState, type MouseEvent, type ReactNode } from 'react'
import { Link, useLocation, useOutlet } from 'react-router'
import { loginUrl } from '../api/session'
import { move, swap } from '../theme/motion'
import { glide, isPlainClick } from './glide'

const pages = [
  { to: '/login', label: 'Inicio' },
  { to: '/catalog', label: 'Trámites' },
  { to: '/about', label: 'Sobre SIGTA' },
] as const

// Kept holds the page it got first, so the page that leaves stays the same
// while it fades out, instead of turning into the next one.
function Kept({ children }: { children: ReactNode }) {
  const [kept] = useState(children)
  return kept
}

// PublicLayout frames the pages a visitor reads without a session: the home
// page, the catalog and the page about SIGTA. The header stays put from one to
// another; the mark under the current link travels to the one opened, and the
// content below gives way to the next one (swap in theme/motion.ts). Its look
// is the one of the landing (landing.module.css, the landing-* variants).
export function PublicLayout() {
  const { pathname } = useLocation()
  // The pages read the person from the outlet; here there is none.
  const outlet = useOutlet(null)
  // A link to the page the person is already on takes them to its top,
  // where it would otherwise do nothing.
  const toTop = (to: string) => (event: MouseEvent) => {
    if (pathname === to && isPlainClick(event)) {
      event.preventDefault()
      glide('top')
    }
  }

  return (
    <Stack variant="landing-page">
      <Group component="header" variant="landing-header">
        <Anchor component={Link} to="/login" aria-label="SIGTA, inicio" onClick={toTop('/login')}>
          <Image variant="landing-logo" src="/brand/sigta-horizontal-descriptor-color.svg" alt="" />
        </Anchor>
        <Group component="nav" variant="landing-nav" aria-label="SIGTA">
          {pages.map((page) => {
            const current = pathname === page.to
            return (
              <Anchor
                key={page.to}
                component={Link}
                to={page.to}
                variant="landing-nav-link"
                aria-current={current ? 'page' : undefined}
                onClick={toTop(page.to)}
              >
                {page.label}
                {current && (
                  <Paper
                    variant="landing-nav-mark"
                    aria-hidden
                    renderRoot={(props) => <motion.span {...props} layoutId="landing-nav-mark" transition={move} />}
                  />
                )}
              </Anchor>
            )
          })}
        </Group>
        <Button component="a" href={loginUrl} variant="landing-login">
          Iniciar sesión
        </Button>
      </Group>

      <AnimatePresence mode="wait" initial={false} onExitComplete={() => window.scrollTo({ top: 0 })}>
        <Stack
          key={pathname}
          component="main"
          variant={pathname === '/login' ? 'landing-main' : 'landing-content'}
          renderRoot={(props) => (
            <motion.main {...props} variants={swap} initial="hidden" animate="shown" exit="gone" />
          )}
        >
          <Kept>{outlet}</Kept>
        </Stack>
      </AnimatePresence>

      <Paper component="footer" variant="landing-footer">
        <Anchor component={Link} to="/login" onClick={toTop('/login')}>
          <Image
            variant="landing-footer-logo"
            src="/brand/sigta-horizontal-descriptor-white.svg"
            alt="SIGTA, Sistema Integral de Gestión de Trámites Académicos"
          />
        </Anchor>
      </Paper>
    </Stack>
  )
}
