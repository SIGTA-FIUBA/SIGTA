import { Box, Group, Paper, Stack, Text } from '@mantine/core'
import { Link } from 'react-router'
import { Icon } from '../Icon/Icon'

export type PendingRowProps = {
  to: string
  // state goes with the link, such as the page to come back to.
  state?: unknown
  // task is what has to be done, as the step or the request names it.
  task: string
  // context is where it belongs: the procedure and, for the staff, its
  // applicant.
  context: string
  // due is the deadline, already written for people, when there is one.
  due?: string
}

// PendingRow is one thing to do, in a list of rows (ProcedureRows): on top
// what has to be done, below the procedure it belongs to and, on the right,
// the deadline when there is one. The whole row is one link, with no box.
export function PendingRow({ to, state, task, context, due }: PendingRowProps) {
  return (
    <Paper
      component={Link}
      to={to}
      state={state}
      viewTransition
      variant="row"
      radius={0}
      px={0}
      py="md"
      td="none"
      c="inherit"
      display="block"
    >
      <Group wrap="nowrap" gap="md" align="flex-start">
        <Stack gap="xxs" flex={1} miw={0}>
          <Text fw={600}>{task}</Text>
          <Text size="sm" c="gray.7" lineClamp={2}>
            {context}
          </Text>
        </Stack>
        {due && (
          <Text size="sm" c="gray.7" ta="right">
            {due}
          </Text>
        )}
        <Box data-chevron aria-hidden>
          <Icon name="chevron-right" size={16} />
        </Box>
      </Group>
    </Paper>
  )
}
