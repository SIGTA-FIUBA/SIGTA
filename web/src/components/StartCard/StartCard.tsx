import { Box, Group, Paper, Stack, Text } from '@mantine/core'
import { Link } from 'react-router'
import { Icon } from '../Icon/Icon'

export type StartCardProps = {
  to: string
  name: string
  description?: string
  // fields and documents are how many the procedure asks for at its start.
  fields: number
  documents: number
}

function count(value: number, one: string, many: string) {
  return `${value} ${value === 1 ? one : many}`
}

// StartCard offers to start one procedure type: its name, what it is in two
// lines at most, and what it asks for; on a phone, where the cards go one
// under the other, only the name and what it asks for. The whole card is one link, a panel
// that answers the pointer (Paper variant="interactive").
export function StartCard({ to, name, description, fields, documents }: StartCardProps) {
  return (
    <Paper
      component={Link}
      to={to}
      variant="interactive"
      withBorder
      p={{ base: 'sm', sm: 'lg' }}
      td="none"
      c="inherit"
      display="flex"
      h="100%"
    >
      <Stack gap="xs" flex={1} justify="space-between">
        <Stack gap="xxs">
          <Text fw={600}>{name}</Text>
          {description && (
            <Text size="sm" c="gray.7" lineClamp={2} visibleFrom="sm">
              {description}
            </Text>
          )}
        </Stack>
        <Group justify="space-between" wrap="nowrap">
          <Text size="xs" c="gray.6">
            {count(fields, 'dato', 'datos')}, {count(documents, 'documento', 'documentos')}
          </Text>
          <Box data-chevron aria-hidden>
            <Icon name="arrow-right" size={16} />
          </Box>
        </Group>
      </Stack>
    </Paper>
  )
}
