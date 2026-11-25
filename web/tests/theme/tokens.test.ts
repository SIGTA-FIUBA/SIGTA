import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { tokens } from '../../src/theme/tokens'

// tailwind.css gives the vendored components the scales of the theme. It
// repeats the breakpoints, which a media query cannot read from a variable,
// and names every shade, so it has to follow the tokens when they change.
const tailwind = readFileSync(`${import.meta.dirname}/../../src/theme/tailwind.css`, 'utf8')

describe('tokens in tailwind.css', () => {
  it('has the breakpoints of the tokens', () => {
    for (const [name, value] of Object.entries(tokens.breakpoints)) {
      expect(tailwind).toContain(`--breakpoint-${name}: ${value};`)
    }
  })

  it('names every shade of every color scale', () => {
    for (const name of Object.keys(tokens.color.scales)) {
      for (let shade = 0; shade < 10; shade++) {
        expect(tailwind).toContain(`--color-${name}-${shade}: var(--mantine-color-${name}-${shade});`)
      }
    }
  })

  it('names every spacing and the three weights', () => {
    for (const name of Object.keys(tokens.spacing)) {
      expect(tailwind).toContain(`--spacing-${name}: var(--mantine-spacing-${name});`)
    }
    expect(tailwind).toContain(`--font-weight-normal: ${tokens.font.weights.regular};`)
    expect(tailwind).toContain(`--font-weight-medium: ${tokens.font.weights.medium};`)
    expect(tailwind).toContain(`--font-weight-semibold: ${tokens.font.weights.bold};`)
  })
})
