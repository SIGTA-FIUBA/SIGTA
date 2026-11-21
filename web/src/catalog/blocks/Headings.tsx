import { Box, Stack, Text, Title } from '@mantine/core'
import { Frame } from './Frame'
import { textStyles } from './textStyleList'

// Headings shows Title order 1 to 3 and the display variant of the landing
// page. Orders 4 to 6 look like order 3.
export function Headings() {
  const orders = [1, 2, 3] as const
  return (
    <Frame>
      <Stack gap="md">
        <Box>
          <Text size="xs" c="dimmed">
            Title variant=&quot;display&quot;: crece con el ancho de pantalla, solo en la pantalla de ingreso
          </Text>
          <Title order={1} variant="display">
            Trámites académicos, sin papeles
          </Title>
        </Box>
        {orders.map((order) => (
          <Box key={order}>
            <Text size="xs" c="dimmed">
              Title order={order}: {textStyles[order - 1]!.name}
            </Text>
            <Title order={order}>Equivalencia entre asignaturas</Title>
          </Box>
        ))}
      </Stack>
    </Frame>
  )
}
