import { type MantineColorsTuple, rem } from '@mantine/core'
import source from './tokens.json'

// Reads tokens.json, the single source of the look. theme.ts builds the Mantine
// theme from these values and the catalog (/design/) shows them, so a change in
// the JSON reaches every screen and the catalog at once.

type Token = { $value: unknown }
type Sizes = Record<'xs' | 'sm' | 'md' | 'lg' | 'xl', string>

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
// Mantine scales its sizes with --mantine-scale only when they come in pixels.
const scaled = (raw: string) => rem(raw.endsWith('rem') ? Number.parseFloat(raw) * 16 : raw)
const stack = (path: string) =>
  (value(path) as string[]).map((family) => (family.includes(' ') ? `"${family}"` : family)).join(', ')
const sizes = (group: string, convert: (raw: string) => string = (raw) => raw): Sizes => ({
  xs: convert(text(`${group}.xs`)),
  sm: convert(text(`${group}.sm`)),
  md: convert(text(`${group}.md`)),
  lg: convert(text(`${group}.lg`)),
  xl: convert(text(`${group}.xl`)),
})
const scale = (path: string) =>
  Array.from({ length: 10 }, (_, index) => text(`${path}.${index}`)) as unknown as MantineColorsTuple

export const tokens = {
  color: {
    brand: { navy: text('color.brand.navy'), cyan: text('color.brand.cyan'), ink: text('color.brand.ink') },
    fiuba: scale('color.fiuba'),
    text: text('color.text'),
    white: text('color.white'),
  },
  primaryColor: text('theme.primaryColor'),
  primaryShade: value('theme.primaryShade') as 7,
  dimmedShade: value('theme.dimmedShade') as number,
  font: {
    sans: stack('font.family.sans'),
    mono: stack('font.family.mono'),
    brand: stack('font.family.brand'),
    headingWeight: text('font.weight.heading'),
    sizes: sizes('font.size', scaled),
    lineHeights: sizes('font.lineHeight'),
  },
  headings: Object.fromEntries(
    (['h1', 'h2', 'h3', 'h4', 'h5', 'h6'] as const).map((order) => [
      order,
      { fontSize: scaled(text(`heading.${order}.fontSize`)), lineHeight: text(`heading.${order}.lineHeight`) },
    ]),
  ) as Record<'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6', { fontSize: string; lineHeight: string }>,
  spacing: sizes('spacing', scaled),
  radius: sizes('radius', scaled),
  defaultRadius: text('radius.default'),
  breakpoints: sizes('breakpoint'),
  component: {
    button: { radius: text('component.button.radius') },
    input: { radius: text('component.input.radius') },
    paper: { radius: text('component.paper.radius') },
    menuLink: { radius: text('component.menuLink.radius') },
    touchTarget: { minHeight: text('component.touchTarget.minHeight') },
  },
}

export { source as tokenSource }
