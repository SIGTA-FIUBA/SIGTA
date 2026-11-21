import {
  Button,
  type CSSVariablesResolver,
  createTheme,
  Input,
  List,
  Modal,
  NavLink,
  Paper,
  rem,
  SegmentedControl,
  Spoiler,
  Text,
  Title,
} from '@mantine/core'
import classes from './interactions.module.css'
import { tokens } from './tokens'
import typography from './typography.module.css'

// The only place where the interface gets its look. Screens use Mantine
// components with theme values (sizes such as "md", colors such as "dimmed"),
// never raw colors or measures; the lint rules in lint/ enforce it. The values
// come from tokens.json, so a design decision is changed there and nowhere else.

// A word with no place to break, such as a file name or a pasted link, breaks
// where it has to instead of running past its box. break-word, unlike
// anywhere, does not narrow the box to fit the word broken up.
const breakLongWords = { root: { overflowWrap: 'break-word' as const } }

export const theme = createTheme({
  colors: { fiuba: tokens.color.fiuba },
  primaryColor: tokens.primaryColor,
  primaryShade: tokens.primaryShade,
  black: tokens.color.text,
  white: tokens.color.white,
  defaultRadius: tokens.defaultRadius,
  radius: tokens.radius,
  spacing: tokens.spacing,
  fontSizes: tokens.font.sizes,
  lineHeights: tokens.font.lineHeights,
  breakpoints: tokens.breakpoints,
  cursorType: 'pointer',
  fontFamily: tokens.font.sans,
  fontFamilyMonospace: tokens.font.mono,
  headings: { fontFamily: tokens.font.sans, fontWeight: tokens.font.headingWeight, sizes: tokens.headings },
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
      defaultProps: { radius: tokens.component.paper.radius },
      classNames: (_theme, props) => ({ root: props.variant === 'interactive' ? classes.interactive : '' }),
    }),
    Input: Input.extend({
      defaultProps: { radius: tokens.component.input.radius },
      classNames: { input: classes.touch },
    }),
    Button: Button.extend({
      defaultProps: { radius: tokens.component.button.radius },
      classNames: { root: classes.touch },
    }),
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
// text needs; tokens.json sets gray 7, which reaches 8.2:1. The other variables
// carry tokens to the CSS modules of this folder.
export const cssVariablesResolver: CSSVariablesResolver = () => ({
  variables: {
    '--sigta-menu-link-radius': `var(--mantine-radius-${tokens.component.menuLink.radius})`,
    '--sigta-touch-target': rem(tokens.component.touchTarget.minHeight),
  },
  light: { '--mantine-color-dimmed': `var(--mantine-color-gray-${tokens.dimmedShade})` },
  dark: {},
})
