import { SimpleGrid } from '@mantine/core'
import { tokens } from '../../theme/tokens'
import { Frame } from './Frame'
import { Swatch } from './Swatch'

const SCALE_NOTES: Record<string, Record<number, string>> = {
  fiuba: {
    0: 'brand.50: selección, avatar',
    5: 'brand.500: cian institucional',
    7: 'action.600: primario',
    8: 'action.700: hover',
  },
  gray: {
    0: 'surface.page: fondo',
    1: 'surface.hover',
    3: 'border.subtle: bordes y divisores',
    4: 'border.control: campos',
    6: 'ink.500: metadatos',
    7: 'ink.600: texto atenuado (dimmed)',
    9: 'ink.900: texto',
  },
  green: { 0: 'success.50', 7: 'success.700' },
  orange: { 0: 'warning.50', 7: 'warning.700' },
  attention: { 0: 'required.50', 7: 'required.700' },
  red: { 0: 'danger.50', 7: 'danger.700' },
}

// ColorScale shows the ten shades of a Mantine color, with their contrast on
// white and the semantic token each one carries.
export function ColorScale({ name }: { name: keyof typeof tokens.color.scales }) {
  const notes = SCALE_NOTES[name] ?? {}
  return (
    <Frame>
      <SimpleGrid cols={{ base: 2, sm: 5 }} spacing="xs">
        {tokens.color.scales[name].map((shade, index) => (
          <Swatch key={shade} color={shade} label={`${name}.${index}`} note={notes[index]} />
        ))}
      </SimpleGrid>
    </Frame>
  )
}
