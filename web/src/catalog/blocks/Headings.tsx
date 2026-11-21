import { Box, Stack, Text, Title } from '@mantine/core'
import { tokenSource, tokens } from '../../theme/tokens'
import { Frame } from './Frame'

// Headings shows Title order 1 to 6 and the display variant.
export function Headings() {
  const orders = [1, 2, 3, 4, 5, 6] as const
  return (
    <Frame>
      <Stack gap="md">
        <Box>
          <Text size="xs" c="dimmed">
            Title variant=&quot;display&quot;: crece con el ancho de pantalla, solo en la portada
          </Text>
          <Title order={1} variant="display">
            Trámites académicos, sin papeles
          </Title>
        </Box>
        {orders.map((order) => {
          const heading = tokenSource.heading[`h${order}`]
          return (
            <Box key={order}>
              <Text size="xs" c="dimmed">
                Title order={order}: {heading.fontSize.$value}, interlineado {heading.lineHeight.$value}, peso{' '}
                {tokens.font.headingWeight}
              </Text>
              <Title order={order}>Equivalencia entre asignaturas</Title>
            </Box>
          )
        })}
      </Stack>
    </Frame>
  )
}
