import { Box, Group, Paper, Stack, Text } from '@mantine/core'
import { Link } from 'react-router'
import { Icon } from '../Icon/Icon'

export type StartCardProps = {
  to: string
  name: string
  description?: string
}

// StartCard offers to start one procedure type: its name and what it is in
// three lines at most; on a phone, where the cards go one under the other,
// only the name. What the procedure asks for is left to its form. The whole
// card is one link, a panel that answers the pointer (Paper
// variant="interactive").
export function StartCard({ to, name, description }: StartCardProps) {
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
      <Stack gap="md" flex={1} miw={0} justify="space-between">
        <Stack gap="xxs">
          <Text fw={600} size="lg">
            {name}
          </Text>
          {description && (
            <Text size="sm" c="gray.7" lineClamp={3} visibleFrom="sm">
              {description}
            </Text>
          )}
        </Stack>
        <Group justify="flex-end" wrap="nowrap" c="fiuba.7">
          <Box data-chevron aria-hidden>
            <Icon name="arrow-right" size={16} />
          </Box>
        </Group>
      </Stack>
    </Paper>
  )
}
