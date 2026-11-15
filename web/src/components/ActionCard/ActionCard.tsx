import { Group, Paper, Stack, Text } from '@mantine/core'
import { Link } from 'react-router'
import { Icon } from '../Icon/Icon'

type ActionCardProps = {
  to: string
  title: string
  description: string
}

// ActionCard is the only box of the content, and it exists only because the
// whole card is one link. A single access is a button, not a card.
export function ActionCard({ to, title, description }: ActionCardProps) {
  return (
    <Paper component={Link} to={to} variant="interactive" withBorder p="lg" td="none" c="inherit">
      <Group justify="space-between" gap="md" wrap="nowrap">
        <Stack gap="xxs">
          <Text fw={700}>{title}</Text>
          <Text size="sm" c="dimmed">
            {description}
          </Text>
        </Stack>
        <Text component="span" c="gray.6" display="flex">
          <Icon name="chevron-right" size={16} />
        </Text>
      </Group>
    </Paper>
  )
}
