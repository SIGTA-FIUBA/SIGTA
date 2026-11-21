import { Box, Code, Group, Stack, Text } from '@mantine/core'
import { tokenSource } from '../../theme/tokens'
import { Frame } from './Frame'
import { SIZES } from './catalogValues'

// Spacing shows the spacing scale as bars.
export function Spacing() {
  return (
    <Frame>
      <Stack gap="xs">
        {SIZES.map((size) => (
          <Group key={size} gap="md" wrap="nowrap">
            <Code w={40}>{size}</Code>
            <Text size="sm" w={80}>
              {tokenSource.spacing[size].$value}
            </Text>
            <Box h={16} bg="fiuba.5" style={{ width: `var(--mantine-spacing-${size})` }} />
          </Group>
        ))}
      </Stack>
    </Frame>
  )
}
