import { SimpleGrid } from '@mantine/core'
import { tokenSource, tokens } from '../../theme/tokens'
import { Frame } from './Frame'
import { Swatch } from './Swatch'

// BrandColors shows the colors of the logo and of the text.
export function BrandColors() {
  const brand = tokenSource.color.brand
  return (
    <Frame>
      <SimpleGrid cols={{ base: 1, sm: 4 }} spacing="xs">
        <Swatch color={brand.navy.$value} label="brand.navy" note={brand.navy.$description} />
        <Swatch color={brand.cyan.$value} label="brand.cyan" note={brand.cyan.$description} />
        <Swatch color={brand.ink.$value} label="brand.ink" note={brand.ink.$description} />
        <Swatch color={tokens.color.text} label="text" note="Texto de la interfaz (Mantine black)." />
      </SimpleGrid>
    </Frame>
  )
}
