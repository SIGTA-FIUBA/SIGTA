import { type MantineColorsTuple, rem } from '@mantine/core'
import source from './tokens.json'

// Reads tokens.json, the single source of the look. theme.ts builds the Mantine
// theme from these values and the catalog (/design/) shows them, so a change in
// the JSON reaches every screen and the catalog at once.

type Token = { $value: unknown }
type Sizes = Record<'xs' | 'sm' | 'md' | 'lg' | 'xl', string>
type Spacing = Sizes & Record<'xxs' | 'xxl' | 'xxxl', string>

const REFERENCE = /^\{([^}]+)\}$/

// The value of a token, following references such as "{radius.default}".
export function value(path: string): unknown {
  let node: unknown = source
  for (const key of path.split('.')) {
    node = (node as Record<string, unknown> | undefined)?.[key]
  }
  if (node === undefined || typeof node !== 'object' || node === null || !('$value' in node)) {
    throw new Error(`tokens.json has no token "${path}"`)
  }
  const raw = (node as Token).$value
  const reference = typeof raw === 'string' ? REFERENCE.exec(raw) : null
  return reference ? value(reference[1]) : raw
}

const text = (path: string) => String(value(path))
// Mantine scales its sizes with --mantine-scale when they come in pixels.
const scaled = (path: string) => rem(text(path))
const stack = (path: string) =>
  (value(path) as string[]).map((family) => (family.includes(' ') ? `"${family}"` : family)).join(', ')
const KEYS = ['xs', 'sm', 'md', 'lg', 'xl'] as const
const sizes = (group: string, convert: (path: string) => string = text): Sizes =>
  Object.fromEntries(KEYS.map((key) => [key, convert(`${group}.${key}`)])) as Sizes
const scale = (path: string) =>
  Array.from({ length: 10 }, (_, index) => text(`${path}.${index}`)) as unknown as MantineColorsTuple
const SCALES = ['fiuba', 'gray', 'green', 'orange', 'red'] as const
const headingFor = (style: string) => ({
  fontSize: scaled(`textStyle.${style}.size`),
  lineHeight: scaled(`textStyle.${style}.lineHeight`),
  fontWeight: text(`textStyle.${style}.weight`),
})
const ms = (path: string) => Number.parseFloat(text(path))

export const tokens = {
  color: {
    brand: { 500: text('color.brand.500'), 50: text('color.brand.50') },
    action: { 600: text('color.action.600'), 700: text('color.action.700') },
    ink: { 900: text('color.ink.900'), 600: text('color.ink.600'), 500: text('color.ink.500') },
    surface: {
      page: text('color.surface.page'),
      raised: text('color.surface.raised'),
      hover: text('color.surface.hover'),
    },
    nav: {
      background: text('color.nav.background'),
      text: text('color.nav.text'),
      muted: text('color.nav.muted'),
      hover: text('color.nav.hover'),
      searchBackground: text('color.nav.searchBackground'),
      searchBorder: text('color.nav.searchBorder'),
      activeBackground: text('color.nav.activeBackground'),
      activeText: text('color.nav.activeText'),
      activeBar: text('color.nav.activeBar'),
      border: text('color.nav.border'),
    },
    border: { subtle: text('color.border.subtle'), control: text('color.border.control') },
    scales: Object.fromEntries(SCALES.map((name) => [name, scale(`color.${name}`)])) as Record<
      (typeof SCALES)[number],
      MantineColorsTuple
    >,
    white: text('color.white'),
    landing: {
      navy: text('color.landing.navy'),
      cyan: text('color.landing.cyan'),
      card: text('color.landing.card'),
      mist: text('color.landing.mist'),
    },
  },
  gradient: {
    login: text('gradient.login'),
    pill: text('gradient.pill'),
    banner: text('gradient.banner'),
    plus: text('gradient.plus'),
    veil: text('gradient.veil'),
  },
  primaryColor: text('theme.primaryColor'),
  primaryShade: value('theme.primaryShade') as 7,
  dimmedShade: value('theme.dimmedShade') as number,
  font: {
    sans: stack('font.family.sans'),
    mono: stack('font.family.mono'),
    brand: stack('font.family.brand'),
    logo: stack('font.family.logo'),
    landing: stack('font.family.landing'),
    // Mantine's bold is the semibold: no component draws a 700.
    weights: {
      regular: text('font.weight.regular'),
      medium: text('font.weight.medium'),
      bold: text('font.weight.semibold'),
    },
    trackingTitle: text('font.tracking.title'),
    trackingDisplay: text('font.tracking.display'),
    trackingSection: text('font.tracking.section'),
    display: { size: text('font.fluid.display.size'), lineHeight: text('font.fluid.display.lineHeight') },
    section: { size: text('font.fluid.section.size'), lineHeight: text('font.fluid.section.lineHeight') },
    figure: { size: text('font.fluid.figure.size'), lineHeight: text('font.fluid.figure.lineHeight') },
    trackingFigure: text('font.tracking.figure'),
    sizes: sizes('font.size', scaled),
    lineHeights: sizes('font.lineHeight', scaled),
  },
  headings: {
    h1: headingFor('title-page'),
    h2: headingFor('title-section'),
    h3: headingFor('title-item'),
    h4: headingFor('title-item'),
    h5: headingFor('title-item'),
    h6: headingFor('title-item'),
  },
  spacing: {
    xxs: scaled('spacing.xxs'),
    ...sizes('spacing', scaled),
    xxl: scaled('spacing.xxl'),
    xxxl: scaled('spacing.xxxl'),
  } as Spacing,
  radius: sizes('radius', scaled),
  defaultRadius: text('radius.default'),
  shadow: text('shadow.overlay'),
  borderWidth: scaled('border.width'),
  focus: { width: scaled('focus.width'), offset: scaled('focus.offset') },
  motion: {
    fast: ms('motion.fast'),
    base: ms('motion.base'),
    easing: `cubic-bezier(${(value('motion.easing') as number[]).join(', ')})`,
    curve: value('motion.easing') as [number, number, number, number],
    nudge: scaled('motion.nudge'),
    press: value('motion.press') as number,
    slow: ms('motion.slow'),
    stagger: ms('motion.stagger'),
    count: ms('motion.count'),
    route: ms('motion.route'),
    rise: Number.parseFloat(text('motion.rise')),
    blur: Number.parseFloat(text('motion.blur')),
    easingOut: value('motion.easingOut') as [number, number, number, number],
    staggerLimit: value('motion.staggerLimit') as number,
    draw: ms('motion.draw'),
    flow: ms('motion.flow'),
    rest: ms('motion.rest'),
  },
  breakpoints: sizes('breakpoint'),
  layout: {
    content: scaled('layout.content'),
    pagePadding: scaled('layout.pagePadding'),
    nav: Number.parseFloat(text('layout.nav')),
    navCompact: Number.parseFloat(text('layout.navCompact')),
    topBar: Number.parseFloat(text('layout.topBar')),
    logo: scaled('layout.logo'),
    signature: scaled('layout.signature'),
    measure: text('layout.measure'),
    iconStroke: Number.parseFloat(text('layout.iconStroke')),
    landing: {
      canvas: value('layout.landing.canvas') as number,
      content: value('layout.landing.content') as number,
      scale: value('layout.landing.scale') as number,
    },
  },
  component: {
    button: { radius: text('component.button.radius'), height: scaled('component.button.height') },
    input: { radius: text('component.input.radius'), height: scaled('component.input.height') },
    card: { radius: text('component.card.radius') },
    navItem: {
      radius: text('component.navItem.radius'),
      height: scaled('component.navItem.height'),
      dot: Number.parseFloat(text('component.navItem.dot')),
      bar: scaled('component.navItem.bar'),
    },
    counter: { height: scaled('component.counter.height') },
    skeleton: { row: scaled('component.skeleton.row'), figures: scaled('component.skeleton.figures') },
    avatar: { size: scaled('component.avatar.size') },
    badge: { height: scaled('component.badge.height') },
    touchTarget: { minHeight: scaled('component.touchTarget.minHeight') },
  },
}

export { source as tokenSource }
