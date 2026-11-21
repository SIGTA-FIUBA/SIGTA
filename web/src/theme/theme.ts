import {
  AppShell,
  Avatar,
  Badge,
  Button,
  Container,
  type CSSVariablesResolver,
  createTheme,
  defaultVariantColorsResolver,
  Input,
  List,
  Menu,
  Modal,
  NavLink,
  Paper,
  Popover,
  rem,
  SegmentedControl,
  Skeleton,
  Spoiler,
  Text,
  Title,
  Tooltip,
  type VariantColorsResolver,
} from '@mantine/core'
import classes from './interactions.module.css'
import { tokens } from './tokens'
import typography from './typography.module.css'

// The only place where the interface gets its look. Screens use Mantine
// components with theme values (sizes such as "md", colors such as "dimmed"),
// never raw colors or measures; the lint rules in lint/ enforce it. The values
// come from tokens.json, so a design decision is changed there and nowhere else.

// Menus, popovers and panels open in motion.base, on the standard curve.
const transitionProps = { duration: tokens.motion.base, timingFunction: tokens.motion.easing }

const shadows = { xs: tokens.shadow, sm: tokens.shadow, md: tokens.shadow, lg: tokens.shadow, xl: tokens.shadow }

// The light variant (badges, avatars, light buttons) writes the 700 shade on
// the 50 tint of the same color, as tokens.json defines every state. Mantine
// would mix the color with transparency instead.
const variantColorResolver: VariantColorsResolver = (input) => {
  const colors = defaultVariantColorsResolver(input)
  if (input.variant !== 'light') {
    return colors
  }
  const name = input.color ?? input.theme.primaryColor
  if (!(name in input.theme.colors)) {
    return colors
  }
  return {
    ...colors,
    background: `var(--mantine-color-${name}-0)`,
    hover: `var(--mantine-color-${name}-1)`,
    color: `var(--mantine-color-${name}-7)`,
  }
}

// Controls of the default size are as tall as layout.control.
const controlHeight = (size: unknown, height: string, variable: string) =>
  size === undefined || size === 'sm' ? { [variable]: height } : {}

// A word with no place to break, such as a file name or a pasted link, breaks
// where it has to instead of running past its box. break-word, unlike
// anywhere, does not narrow the box to fit the word broken up.
const breakLongWords = { root: { overflowWrap: 'break-word' as const } }

export const theme = createTheme({
  colors: tokens.color.scales,
  primaryColor: tokens.primaryColor,
  primaryShade: tokens.primaryShade,
  black: tokens.color.ink[900],
  white: tokens.color.white,
  defaultRadius: tokens.defaultRadius,
  radius: tokens.radius,
  spacing: tokens.spacing,
  shadows,
  fontSizes: tokens.font.sizes,
  lineHeights: tokens.font.lineHeights,
  fontWeights: tokens.font.weights,
  breakpoints: tokens.breakpoints,
  cursorType: 'pointer',
  // Pressed, a control shrinks to motion.press; with reduced motion Mantine
  // drops its own transitions.
  activeClassName: classes.pressed,
  respectReducedMotion: true,
  fontFamily: tokens.font.sans,
  fontFamilyMonospace: tokens.font.mono,
  headings: { fontFamily: tokens.font.sans, fontWeight: tokens.font.weights.bold, sizes: tokens.headings },
  variantColorResolver,
  other: {
    // The height of a document opened in the page, below the header of its
    // full screen modal.
    documentHeight: 'calc(100dvh - 9rem)',
  },
  components: {
    // An item of the side navigation (variant="menu"), on white: ink.600 at
    // rest, surface.hover under the pointer and the current page in brand.50
    // with a bar on its left (interactions.module.css). On a tablet, variant
    // "menu-icon" shows only the icon.
    NavLink: NavLink.extend({
      classNames: (_theme, props) => ({
        root:
          props.variant === 'menu'
            ? `${classes.touch} ${classes.navItem}`
            : props.variant === 'menu-icon'
              ? `${classes.touch} ${classes.navItem} ${classes.navIcon}`
              : classes.touch,
        label: props.variant === 'menu' ? classes.navLabel : '',
        description: props.variant === 'menu' ? classes.navDescription : '',
      }),
    }),
    // The side navigation is white with a border on its right; the page is
    // gray, except on the landing page (variant="landing").
    AppShell: AppShell.extend({
      defaultProps: { transitionDuration: tokens.motion.base, transitionTimingFunction: tokens.motion.easing },
      classNames: (_theme, props) => ({
        navbar: classes.sideNav,
        main: props.variant === 'landing' ? '' : classes.page,
      }),
    }),
    // A panel with a border and a background of its own is a control: the
    // action card and the rows use variant="interactive", which answers the
    // pointer; the container of rows uses variant="rows" and the summary of
    // figures variant="figures".
    Paper: Paper.extend({
      defaultProps: { radius: tokens.component.card.radius },
      classNames: (_theme, props) => ({
        root:
          props.variant === 'interactive'
            ? classes.interactive
            : props.variant === 'rows'
              ? classes.rows
              : props.variant === 'figures'
                ? classes.figures
                : '',
      }),
    }),
    // The steps of a procedure: an ordered list in columns (StepList).
    List: List.extend({
      classNames: (_theme, props) =>
        props.variant === 'steps' ? { root: classes.steps, item: classes.stepsItem } : {},
      styles: { item: breakLongWords.root },
    }),
    Input: Input.extend({
      defaultProps: { radius: tokens.component.input.radius },
      vars: (_theme, props) => ({
        wrapper: controlHeight(props.size, tokens.component.input.height, '--input-height'),
      }),
      classNames: { input: classes.touch },
    }),
    Button: Button.extend({
      defaultProps: { radius: tokens.component.button.radius },
      vars: (_theme, props) => ({
        root: controlHeight(props.size, tokens.component.button.height, '--button-height'),
      }),
      classNames: { root: `${classes.touch} ${classes.control}` },
    }),
    // A state reads in its text: no capitals, caption at 600, 700 on 50.
    // variant="counter" is the count of the inbox in the menu.
    Badge: Badge.extend({
      defaultProps: { variant: 'light' },
      vars: (_theme, props) => ({
        root:
          props.variant === 'counter'
            ? {
                '--badge-height': tokens.component.counter.height,
                '--badge-fz': 'var(--mantine-font-size-xs)',
                '--badge-padding-x': 'var(--mantine-spacing-xxs)',
                '--badge-bg': 'var(--mantine-primary-color-filled)',
                '--badge-color': 'var(--mantine-color-white)',
                '--badge-bd': '0',
              }
            : props.size === undefined || props.size === 'md'
              ? {
                  '--badge-height': tokens.component.badge.height,
                  '--badge-fz': 'var(--mantine-font-size-xs)',
                  '--badge-padding-x': 'var(--mantine-spacing-xs)',
                }
              : {},
      }),
      classNames: (_theme, props) => ({
        root: props.variant === 'counter' ? `${classes.badge} ${classes.counter}` : classes.badge,
      }),
    }),
    Avatar: Avatar.extend({
      defaultProps: { color: tokens.primaryColor, variant: 'light' },
      vars: () => ({ root: { '--avatar-size': tokens.component.avatar.size } }),
      classNames: { root: classes.avatar },
    }),
    Container: Container.extend({ defaultProps: { size: 'content' } }),
    SegmentedControl: SegmentedControl.extend({ classNames: { control: classes.touch } }),
    Menu: Menu.extend({ defaultProps: { shadow: 'md', transitionProps } }),
    Popover: Popover.extend({ defaultProps: { shadow: 'md', transitionProps } }),
    Tooltip: Tooltip.extend({ defaultProps: { openDelay: 300, transitionProps } }),
    // Loading shows the shape of what comes, never a spinner.
    Skeleton: Skeleton.extend({ classNames: { root: classes.skeleton } }),
    // Titles of 20px or more are tighter; the headline of the landing page
    // grows with the screen.
    Text: Text.extend({ styles: breakLongWords }),
    Title: Title.extend({
      styles: breakLongWords,
      classNames: (_theme, props) => ({
        root: props.variant === 'display' ? `${typography.title} ${typography.display}` : typography.title,
      }),
    }),
    // A long title, such as the name of a file, breaks instead of pushing the
    // close button out of the screen.
    Modal: Modal.extend({ styles: { title: { ...breakLongWords.root, minWidth: 0 } } }),
    // A long text shows its first lines and opens the rest on request.
    Spoiler: Spoiler.extend({ defaultProps: { maxHeight: 120 } }),
  },
})

// Dimmed text (c="dimmed") is the secondary text, gray 7 (ink.600); Mantine
// would use gray 6, which here is the metadata (ink.500). The hover of
// default controls is surface.hover.
// The other variables carry tokens to the CSS modules of this folder and to
// the screens, which name them instead of writing a measure.
export const cssVariablesResolver: CSSVariablesResolver = () => ({
  variables: {
    '--container-size-content': tokens.layout.content,
    '--sigta-measure': tokens.layout.measure,
    '--sigta-logo-height': tokens.layout.logo,
    '--sigta-signature-height': tokens.layout.signature,
    '--sigta-nav-item-height': tokens.component.navItem.height,
    '--sigta-nav-item-radius': `var(--mantine-radius-${tokens.component.navItem.radius})`,
    '--sigta-nav-background': tokens.color.nav.background,
    '--sigta-nav-text': tokens.color.nav.text,
    '--sigta-nav-hover': tokens.color.nav.hover,
    '--sigta-nav-active-background': tokens.color.nav.activeBackground,
    '--sigta-nav-active-text': tokens.color.nav.activeText,
    '--sigta-nav-active-bar': tokens.color.nav.activeBar,
    '--sigta-nav-border': tokens.color.nav.border,
    '--sigta-nav-bar': tokens.component.navItem.bar,
    '--sigta-counter-height': tokens.component.counter.height,
    '--sigta-row-height': tokens.component.skeleton.row,
    '--sigta-figures-height': tokens.component.skeleton.figures,
    '--sigta-tracking-title': tokens.font.trackingTitle,
    '--sigta-tracking-display': tokens.font.trackingDisplay,
    '--sigta-display-size': tokens.font.display.size,
    '--sigta-display-line-height': tokens.font.display.lineHeight,
    '--sigta-border-width': tokens.borderWidth,
    '--sigta-duration-fast': `${tokens.motion.fast}ms`,
    '--sigta-duration-base': `${tokens.motion.base}ms`,
    '--sigta-ease': tokens.motion.easing,
    '--sigta-nudge': tokens.motion.nudge,
    '--sigta-press': String(tokens.motion.press),
    '--sigta-focus-width': tokens.focus.width,
    '--sigta-focus-offset': tokens.focus.offset,
    '--sigta-touch-target': rem(tokens.component.touchTarget.minHeight),
  },
  light: {
    '--mantine-color-dimmed': `var(--mantine-color-gray-${tokens.dimmedShade})`,
    '--mantine-color-default-hover': 'var(--mantine-color-gray-1)',
  },
  dark: {},
})
