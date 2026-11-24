import { Kbd, Stack } from '@mantine/core'

type SearchShortcutProps = {
  // keycap draws the keys as a key (Kbd), for a field that opens the search.
  keycap?: boolean
}

// SearchShortcut is the shortcut written next to the button that opens the
// search: Cmd K on a Mac, Ctrl K elsewhere.
export function SearchShortcut({ keycap = false }: SearchShortcutProps) {
  const mac = typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform)
  const keys = mac ? '⌘ K' : 'Ctrl K'
  return keycap ? (
    <Kbd size="sm" aria-hidden>
      {keys}
    </Kbd>
  ) : (
    <Stack gap={0} data-shortcut aria-hidden>
      {keys}
    </Stack>
  )
}
