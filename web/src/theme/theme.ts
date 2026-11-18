import {
  Button,
  type CSSVariablesResolver,
  createTheme,
  Input,
  List,
  Modal,
  type MantineColorsTuple,
  NavLink,
  Paper,
  SegmentedControl,
  Spoiler,
  Text,
  Title,
} from '@mantine/core'
import classes from './interactions.module.css'
import typography from './typography.module.css'

// The only place where the interface gets its look. Screens use Mantine
// components with theme values (sizes such as "md", colors such as "dimmed"),
// never raw colors or measures; the lint rules in lint/ enforce it.

// A word with no place to break, such as a file name or a pasted link, breaks
// where it has to instead of running past its box. break-word, unlike
// anywhere, does not narrow the box to fit the word broken up.
const breakLongWords = { root: { overflowWrap: 'break-word' as const } }

// The FIUBA light blue (#009fe2, index 5) is the color of the logo. On white
// it does not reach the contrast text needs (2.98:1), so buttons and links use
// index 7 (#0074a6, 5.19:1) and hovers index 8 (#005f88, 7.02:1).
const fiuba: MantineColorsTuple = [
  '#e5f6fd',
  '#cdebfa',
  '#9bd6f4',
  '#64c0ee',
  '#38ade8',
  '#009fe2',
  '#0089c4',
  '#0074a6',
  '#005f88',
  '#004a6b',
]

export const theme = createTheme({
  colors: { fiuba },
  primaryColor: 'fiuba',
  primaryShade: 7,
  // The near black of the logo, for the text.
  black: '#121516',
  defaultRadius: 'sm',
  cursorType: 'pointer',
  fontFamily: 'system-ui, -apple-system, "Segoe UI", Roboto, sans-serif',
  headings: { fontWeight: '600' },
  other: {
    // The height of a document opened in the page, below the header of its
    // full screen modal.
    documentHeight: 'calc(100dvh - 9rem)',
  },
  components: {
    // The side menu (variant="menu") sits on the logo's light blue: each link
    // is a white box with dark text, and the current page is filled in dark
    // blue with white text.
    NavLink: NavLink.extend({
      vars: (_theme, props) => ({
        root:
          props.variant === 'menu'
            ? {
                '--nl-bg': 'var(--mantine-color-fiuba-8)',
                '--nl-color': 'var(--mantine-color-white)',
                '--nl-hover': 'var(--mantine-color-fiuba-9)',
              }
            : {},
        children: {},
      }),
      classNames: (_theme, props) => ({
        root: props.variant === 'menu' ? `${classes.touch} ${classes.menuLink}` : classes.touch,
      }),
    }),
    // A panel that is itself a control, such as a summary that filters the
    // list, uses variant="interactive" and answers the pointer.
    Paper: Paper.extend({
      classNames: (_theme, props) => ({ root: props.variant === 'interactive' ? classes.interactive : '' }),
    }),
    Input: Input.extend({ classNames: { input: classes.touch } }),
    Button: Button.extend({ classNames: { root: classes.touch } }),
    SegmentedControl: SegmentedControl.extend({ classNames: { control: classes.touch } }),
    Text: Text.extend({ styles: breakLongWords }),
    // The headline of the landing page grows with the screen.
    Title: Title.extend({
      styles: breakLongWords,
      classNames: (_theme, props) => ({ root: props.variant === 'display' ? typography.display : '' }),
    }),
    List: List.extend({ styles: { item: breakLongWords.root } }),
    // A long title, such as the name of a file, breaks instead of pushing the
    // close button out of the screen.
    Modal: Modal.extend({ styles: { title: { ...breakLongWords.root, minWidth: 0 } } }),
    // A long text shows its first lines and opens the rest on request.
    Spoiler: Spoiler.extend({ defaultProps: { maxHeight: 120 } }),
  },
})

// Mantine's dimmed text (gray 6) reads at 3.3:1 on white, below the 4.5:1 that
// text needs; gray 7 reaches 8.2:1.
export const cssVariablesResolver: CSSVariablesResolver = () => ({
  variables: {},
  light: { '--mantine-color-dimmed': 'var(--mantine-color-gray-7)' },
  dark: {},
})
