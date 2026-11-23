import { Anchor, Group, Image } from '@mantine/core'
import type { ReactNode } from 'react'
import { Link } from 'react-router'

type AppHeaderProps = {
  // start goes before the logo, such as the button that opens the menu.
  start?: ReactNode
  // end goes on the right, such as the login.
  end?: ReactNode
}

// AppHeader is the white top bar of the screens without the side navigation:
// the SIGTA logo on the left, as a link to the start page, and the actions on
// the right. With a session it only shows on a phone, to open the menu.
export function AppHeader({ start, end }: AppHeaderProps) {
  return (
    <Group h="100%" px={{ base: 'md', sm: 'lg' }} justify="space-between" wrap="nowrap">
      <Group h="100%" gap="sm" wrap="nowrap">
        {start}
        <Anchor component={Link} to="/" display="flex" aria-label="SIGTA, ir al inicio">
          <Image src="/brand/sigta-horizontal-color.svg" alt="" h="var(--sigta-logo-height)" w="auto" />
        </Anchor>
      </Group>
      {end && (
        <Group gap="sm" wrap="nowrap">
          {end}
        </Group>
      )}
    </Group>
  )
}
