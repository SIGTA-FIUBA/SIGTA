import { createTheme, List, Spoiler, Text, Title } from '@mantine/core'

// A word with no place to break, such as a file name or a pasted link, breaks
// where it has to instead of running past its box. break-word, unlike
// anywhere, does not narrow the box to fit the word broken up.
const breakLongWords = { root: { overflowWrap: 'break-word' as const } }

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
  components: {
    Text: Text.extend({ styles: breakLongWords }),
    Title: Title.extend({ styles: breakLongWords }),
    List: List.extend({ styles: { item: breakLongWords.root } }),
    // A long text shows its first lines and opens the rest on request.
    Spoiler: Spoiler.extend({ defaultProps: { maxHeight: 120 } }),
  },
})
