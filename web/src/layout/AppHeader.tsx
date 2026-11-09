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

// AppHeader is the same top bar on every screen: the faculty's logo and the
// system's name on the left, as a link to the start page, and the actions on
// the right.
export function AppHeader({ start, end }: AppHeaderProps) {
  return (
    <Group h="100%" px="md" justify="space-between" wrap="nowrap">
      <Group h="100%" gap="sm" wrap="nowrap">
        {start}
        <Anchor
          component={Link}
          to="/"
          h="100%"
          display="flex"
          underline="never"
          c="inherit"
          aria-label="SIGTA, ir al inicio"
        >
          <Group h="100%" gap="sm" wrap="nowrap">
            <Image src={logo} alt="" h="70%" w="auto" fit="contain" />
            <Divider orientation="vertical" my="md" visibleFrom="xs" />
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
