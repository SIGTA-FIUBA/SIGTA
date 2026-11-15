import { Anchor, Divider, Group, Image, Text } from '@mantine/core'
import type { ReactNode } from 'react'
import { Link } from 'react-router'
import logo from '../assets/fiuba-logo.png'

type AppHeaderProps = {
  // start goes before the logo, such as the button that opens the menu.
  start?: ReactNode
  // end goes on the right, such as the account menu or the login.
  end: ReactNode
}

// AppHeader is the same top bar on every screen: the SIGTA signature on the
// left (the faculty's logo, a divider and the system's name), as a link to the
// start page, and the actions on the right.
export function AppHeader({ start, end }: AppHeaderProps) {
  return (
    <Group h="100%" px="lg" justify="space-between" wrap="nowrap">
      <Group h="100%" gap="sm" wrap="nowrap">
        {start}
        <Anchor component={Link} to="/" display="flex" underline="never" c="inherit" aria-label="SIGTA, ir al inicio">
          <Group gap="md" wrap="nowrap">
            <Image src={logo} alt="" w="var(--sigta-logo-width)" h="auto" fit="contain" />
            <Divider orientation="vertical" h="var(--sigta-top-bar-divider)" visibleFrom="xs" />
            <Text fw={700} size="lg" visibleFrom="xs">
              SIGTA
            </Text>
          </Group>
        </Anchor>
      </Group>
      <Group gap="sm" wrap="nowrap">
        {end}
      </Group>
    </Group>
  )
}
