import { Box, Flex, Group, Paper, Stack, Text } from '@mantine/core'
import { Children, type ReactNode } from 'react'
import { Link } from 'react-router'
import { StatusBadge, type StatusTone } from '../StatusBadge/StatusBadge'

export type ProcedureRowProps = {
  to: string
  // state goes with the link, such as the page to come back to.
  state?: unknown
  name: string
  // detail says who or what, such as the applicant, under the name.
  detail?: string
  status: { tone: StatusTone; label: string }
  nextStep: string
  // date is the last news of the procedure, already written for people.
  date: string
}

// ProcedureRow is one procedure in a list: its name, its state, the step that
// comes next and the date of the last news. The whole row is one link.
export function ProcedureRow({ to, state, name, detail, status, nextStep, date }: ProcedureRowProps) {
  return (
    <Paper
      component={Link}
      to={to}
      state={state}
      variant="interactive"
      radius={0}
      p="md"
      td="none"
      c="inherit"
      display="block"
    >
      {/* On a narrow screen the state and the date wrap under the name. */}
      <Group justify="space-between" align="flex-start" gap="md">
        <Stack gap="xxs" miw={0}>
          <Text fw={700}>{name}</Text>
          {detail && (
            <Text size="sm" c="dimmed">
              {detail}
            </Text>
          )}
          <Text size="sm" c="dimmed">
            Próximo paso: {nextStep}
          </Text>
        </Stack>
        <Flex
          direction={{ base: 'row', sm: 'column' }}
          align={{ base: 'center', sm: 'flex-end' }}
          gap={{ base: 'xs', sm: 'xxs' }}
        >
          <StatusBadge tone={status.tone}>{status.label}</StatusBadge>
          <Text size="xs" c="gray.6">
            {date}
          </Text>
        </Flex>
      </Group>
    </Paper>
  )
}

// ProcedureRows holds the rows in a single container, with a divider between
// them.
export function ProcedureRows({ label, children }: { label: string; children: ReactNode[] }) {
  return (
    <Paper withBorder variant="rows" component="ul" aria-label={label} p={0} m={0}>
      {Children.map(children, (row) => (
        <Box component="li">{row}</Box>
      ))}
    </Paper>
  )
}
