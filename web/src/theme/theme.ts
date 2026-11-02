import { createTheme } from '@mantine/core'

// The only place where the interface gets its look. Screens use Mantine
// components with theme values (sizes such as "md", colors such as "dimmed"),
// never raw colors or measures; the lint rules in lint/ enforce it.
export const theme = createTheme({
  // To be replaced by the institutional palette once the visual identity is
  // defined.
  primaryColor: 'blue',
  defaultRadius: 'sm',
  cursorType: 'pointer',
  fontFamily: 'system-ui, -apple-system, "Segoe UI", Roboto, sans-serif',
  headings: { fontWeight: '600' },
})
