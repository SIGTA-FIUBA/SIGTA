import { Group, Paper, Stack, Text, ThemeIcon } from '@mantine/core'
import { Link } from 'react-router'
import { Icon } from '../Icon/Icon'

export type PendingCardProps = {
  to: string
  // state goes with the link, such as the page to come back to.
  state?: unknown
  // task is what has to be done, as the request names it.
  task: string
  // context is the procedure it belongs to.
  context: string
  // due is the deadline, already written for people, when there is one.
  due?: string
}

// PendingCard is one thing an applicant has to do, where it cannot be missed:
// a panel with a mark in the action color, what has to be done, the procedure it belongs
// to and the deadline. The whole card is one link to answer it, a panel that
// answers the pointer (Paper variant="interactive").
export function PendingCard({ to, state, task, context, due }: PendingCardProps) {
  return (
    <Paper
      component={Link}
      to={to}
      state={state}
      viewTransition
      variant="interactive"
      withBorder
      p="lg"
      td="none"
      c="inherit"
      display="block"
      h="100%"
    >
      <Group wrap="nowrap" gap="md" align="flex-start" h="100%">
        <ThemeIcon variant="light" color="fiuba" radius="xl" size="lg" aria-hidden>
          <Icon name="file-pencil" size={20} />
        </ThemeIcon>
        <Stack gap="xs" flex={1} miw={0} justify="space-between" h="100%">
          <Stack gap="xxs">
            <Text fw={600}>{task}</Text>
            <Text size="sm" c="gray.7">
              {context}
            </Text>
          </Stack>
          {due && (
            <Text size="sm" c="fiuba.7" fw={500}>
              {due}
            </Text>
          )}
        </Stack>
      </Group>
    </Paper>
  )
}
