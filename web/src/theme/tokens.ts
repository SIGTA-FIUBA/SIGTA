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
    border: { subtle: text('color.border.subtle'), control: text('color.border.control') },
    scales: Object.fromEntries(SCALES.map((name) => [name, scale(`color.${name}`)])) as Record<
      (typeof SCALES)[number],
      MantineColorsTuple
    >,
    white: text('color.white'),
  },
  primaryColor: text('theme.primaryColor'),
  primaryShade: value('theme.primaryShade') as 7,
  dimmedShade: value('theme.dimmedShade') as number,
  font: {
    sans: stack('font.family.sans'),
    mono: stack('font.family.mono'),
    brand: stack('font.family.brand'),
    weights: { regular: text('font.weight.regular'), medium: text('font.weight.bold'), bold: text('font.weight.bold') },
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
  focus: { width: scaled('focus.width'), offset: scaled('focus.offset') },
  breakpoints: sizes('breakpoint'),
  layout: {
    content: scaled('layout.content'),
    nav: Number.parseFloat(text('layout.nav')),
    navCompact: Number.parseFloat(text('layout.navCompact')),
    topBar: Number.parseFloat(text('layout.topBar')),
    topBarDivider: scaled('layout.topBarDivider'),
    logo: scaled('layout.logo'),
    measure: text('layout.measure'),
    iconStroke: Number.parseFloat(text('layout.iconStroke')),
  },
  component: {
    button: { radius: text('component.button.radius'), height: scaled('component.button.height') },
    input: { radius: text('component.input.radius'), height: scaled('component.input.height') },
    card: { radius: text('component.card.radius') },
    navItem: {
      radius: text('component.navItem.radius'),
      height: scaled('component.navItem.height'),
      marker: scaled('component.navItem.marker'),
      dot: Number.parseFloat(text('component.navItem.dot')),
    },
    avatar: { size: scaled('component.avatar.size') },
    badge: { height: scaled('component.badge.height') },
    touchTarget: { minHeight: scaled('component.touchTarget.minHeight') },
  },
}

export { source as tokenSource }
