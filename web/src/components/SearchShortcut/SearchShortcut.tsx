import { Stack } from '@mantine/core'

// SearchShortcut is the shortcut written next to the button that opens the
// search: Cmd K on a Mac, Ctrl K elsewhere.
export function SearchShortcut() {
  const mac = typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform)
  return (
    <Stack gap={0} data-shortcut aria-hidden>
      {mac ? '⌘ K' : 'Ctrl K'}
    </Stack>
  )
}
