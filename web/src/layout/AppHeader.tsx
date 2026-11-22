import { Anchor, Container, Group, Image } from '@mantine/core'
import type { ReactNode } from 'react'
import { Link } from 'react-router'
import { BrandSignature } from '../components/BrandSignature/BrandSignature'

type AppHeaderProps = {
  // start goes before the logo, such as the button that opens the menu.
  start?: ReactNode
  // end goes on the right, such as the login.
  end?: ReactNode
  // signature shows the joint signature with FIUBA instead of the SIGTA logo,
  // where SIGTA presents itself.
  signature?: boolean
  // contained keeps the bar on the width of the content, so the logo and the
  // actions line up with the page below, as on the landing page.
  contained?: boolean
}

// AppHeader is the white top bar of the screens without the side navigation:
// the SIGTA logo (or the joint signature) on the left, as a link to the start page, and the actions on
// the right. With a session it only shows on a phone, to open the menu.
export function AppHeader({ start, end, signature = false, contained = false }: AppHeaderProps) {
  const bar = (
    <Group h="100%" px={contained ? undefined : { base: 'md', sm: 'lg' }} justify="space-between" wrap="nowrap">
      <Group h="100%" gap="sm" wrap="nowrap">
        {start}
        <Anchor component={Link} to="/" display="flex" aria-label="SIGTA, ir al inicio">
          {signature ? (
            <BrandSignature compact />
          ) : (
            <Image src="/brand/sigta-horizontal-color.svg" alt="" h="var(--sigta-logo-height)" w="auto" />
          )}
        </Anchor>
      </Group>
      {end && (
        <Group gap="sm" wrap="nowrap">
          {end}
        </Group>
      )}
    </Group>
  )
  return contained ? <Container h="100%">{bar}</Container> : bar
}
