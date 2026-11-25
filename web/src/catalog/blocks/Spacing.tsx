import { Box, Code, Group, Stack, Text } from '@mantine/core'
import { tokenSource } from '../../theme/tokens'
import { Frame } from './Frame'

const SPACING = ['xxs', 'xs', 'sm', 'md', 'lg', 'xl', 'xxl', 'xxxl'] as const

// Spacing shows the spacing scale as bars.
export function Spacing() {
  return (
    <Frame>
      <Stack gap="xs">
        {SPACING.map((size) => (
          <Group key={size} gap="md" wrap="nowrap">
            <Code w={48}>{size}</Code>
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
