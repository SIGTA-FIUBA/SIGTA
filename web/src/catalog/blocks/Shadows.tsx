import { Code, Paper, SimpleGrid } from '@mantine/core'
import { Frame } from './Frame'
import { SIZES } from './catalogValues'

// Shadows shows each shadow on a panel.
export function Shadows() {
  return (
    <Frame>
      <SimpleGrid cols={{ base: 2, sm: 5 }} spacing="lg">
        {SIZES.map((size) => (
          <Paper key={size} shadow={size} p="md">
            <Code>{size}</Code>
          </Paper>
        ))}
      </SimpleGrid>
    </Frame>
  )
}
