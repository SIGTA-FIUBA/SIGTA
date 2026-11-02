import { MantineProvider } from '@mantine/core'
import { render as testingLibraryRender } from '@testing-library/react'
import type { ReactNode } from 'react'
import { theme } from '../theme/theme'

// render mounts the UI inside the application's theme. env="test" turns off
// transitions and portals, so menus open synchronously.
export function render(ui: ReactNode) {
  return testingLibraryRender(ui, {
    wrapper: ({ children }) => (
      <MantineProvider theme={theme} env="test">
        {children}
      </MantineProvider>
    ),
  })
}
