import { Paper, Text } from '@mantine/core'
import { Frame } from './Frame'

// Shadows shows the only shadow, on what floats.
export function Shadows() {
  return (
    <Frame>
      <Paper shadow="md" p="md" maw={240}>
        <Text size="sm">Un menú desplegado</Text>
      </Paper>
    </Frame>
  )
}
