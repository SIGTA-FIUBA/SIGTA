import { Badge, Code, Group, Paper, SimpleGrid, Text } from '@mantine/core'
import { tokenSource, tokens } from '../../theme/tokens'
import { Frame } from './Frame'
import { SIZES } from './catalogValues'

// Radii shows each radius on a panel and marks the default one.
export function Radii() {
  return (
    <Frame>
      <SimpleGrid cols={{ base: 2, sm: 5 }} spacing="sm">
        {SIZES.map((size) => (
          <Paper key={size} withBorder radius={size} p="md" bg="fiuba.0">
            <Group gap="xs">
              <Code>{size}</Code>
              {size === tokens.defaultRadius && <Badge size="xs">por defecto</Badge>}
            </Group>
            <Text size="xs" mt="xs">
              {tokenSource.radius[size].$value}
            </Text>
          </Paper>
        ))}
      </SimpleGrid>
    </Frame>
  )
}
