import { Box, MantineProvider } from '@mantine/core'
import { Unstyled } from '@storybook/addon-docs/blocks'
import type { ReactNode } from 'react'
import { cssVariablesResolver, theme } from '../../theme/theme'

// Blocks of the catalog's documentation pages (the .mdx files). They draw the
// tokens from tokens.json with the app's theme, so the pages cannot drift from
// the code. Only the catalog uses them; showing a raw value is their purpose,
// which is why .oxlintrc.json lets this folder write styles. Each block
// is a file of its own, named after it.

// Frame puts a block inside the app's theme, away from the styles of the
// documentation page.
export function Frame({ children }: { children: ReactNode }) {
  return (
    <Unstyled>
      <MantineProvider theme={theme} cssVariablesResolver={cssVariablesResolver} forceColorScheme="light">
        <Box py="sm">{children}</Box>
      </MantineProvider>
    </Unstyled>
  )
}
