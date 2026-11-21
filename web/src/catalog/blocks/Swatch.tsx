import { Badge, Box, Paper, Stack, Text } from '@mantine/core'
import { tokens } from '../../theme/tokens'
import { contrast } from './catalogValues'

export function Swatch({ color, label, note }: { color: string; label: string; note?: string }) {
  const onWhite = contrast(color, '#ffffff')
  const ink = onWhite >= 4.5 ? '#ffffff' : tokens.color.text
  return (
    <Paper withBorder radius="sm" style={{ overflow: 'hidden' }}>
      <Box p="sm" style={{ background: color, color: ink, minHeight: 72 }}>
        <Text fw={600} size="sm">
          {label}
        </Text>
        <Text size="xs" ff="monospace">
          {color.toLowerCase()}
        </Text>
      </Box>
      <Stack gap={4} p="xs">
        <Text size="xs">
          {onWhite.toFixed(2)}:1 sobre blanco{' '}
          {onWhite >= 4.5 ? (
            <Badge size="xs">texto</Badge>
          ) : onWhite >= 3 ? (
            <Badge size="xs" color="gray">
              gráficos
            </Badge>
          ) : null}
        </Text>
        {note && (
          <Text size="xs" c="dimmed">
            {note}
          </Text>
        )}
      </Stack>
    </Paper>
  )
}
