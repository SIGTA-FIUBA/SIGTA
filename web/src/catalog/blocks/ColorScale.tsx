import { DEFAULT_THEME, SimpleGrid } from '@mantine/core'
import { tokens } from '../../theme/tokens'
import { Frame } from './Frame'
import { Swatch } from './Swatch'

// ColorScale shows the ten shades of a Mantine color, with their contrast on white.
export function ColorScale({ name }: { name: string }) {
  const shades = name === 'fiuba' ? tokens.color.fiuba : DEFAULT_THEME.colors[name]
  const notes: Record<number, string> =
    name === 'fiuba'
      ? { 5: 'Celeste de la marca, fondos', 7: 'Primario: botones y links', 8: 'Hover del primario' }
      : name === 'gray'
        ? { 0: 'Fondo de las páginas', 7: 'Texto atenuado (dimmed)' }
        : {}
  return (
    <Frame>
      <SimpleGrid cols={{ base: 2, sm: 5 }} spacing="xs">
        {shades.map((shade, index) => (
          <Swatch key={shade} color={shade} label={`${name}.${index}`} note={notes[index]} />
        ))}
      </SimpleGrid>
    </Frame>
  )
}
