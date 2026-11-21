import { Box, Flex, Group, Paper, Stack, Text } from '@mantine/core'
import { Link } from 'react-router'
import { Icon } from '../Icon/Icon'
import { StatusBadge, type StatusTone } from '../StatusBadge/StatusBadge'

export type ProcedureRowProps = {
  to: string
  // state goes with the link, such as the page to come back to.
  state?: unknown
  // type is the name of the procedure type.
  type: string
  // applicant is left out in the lists of the person's own procedures.
  applicant?: string
  // step is the open step, or what took its place once closed.
  step: { name: string }
  // area is who acts now, when the list knows it.
  area?: string
  // note is something the person has to do, such as a correction.
  note?: string
  status: { tone: StatusTone; label: string }
  // date is the last news, already written for people.
  date: string
}

// ProcedureRow is one procedure in a list, in four zones that answer what it
// is, whose it is, at which stage it is and who acts now, with its state and
// its last news. The whole row is one link.
export function ProcedureRow({ to, state, type, applicant, step, area, note, status, date }: ProcedureRowProps) {
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
      <Group wrap="nowrap" gap="md">
        {/* On a narrow screen the zones stack; from md on they share the row. */}
        <Flex
          direction={{ base: 'column', md: 'row' }}
          align={{ base: 'stretch', md: 'center' }}
          gap={{ base: 'sm', md: 'lg' }}
          flex={1}
          miw={0}
        >
          <Stack gap="xxs" miw={0} flex={{ md: 3 }}>
            {/* A cut name stays readable on hover. */}
            <Text fw={600} truncate title={type}>
              {type}
            </Text>
            {applicant && (
              <Text size="sm" c="gray.7">
                {applicant}
              </Text>
            )}
            {note && (
              <Text size="sm" fw={600} c="orange.7">
                {note}
              </Text>
            )}
          </Stack>
          <Stack gap="xs" miw={0} flex={{ md: 2 }}>
            <Text size="xs" c="gray.7">
              {step.name}
            </Text>
          </Stack>
          <Stack gap="xxs" align="flex-start" miw={0} flex={{ md: 2 }}>
            <StatusBadge tone={status.tone}>{status.label}</StatusBadge>
            {area && (
              <Text size="xs" c="gray.7">
                A cargo de {area}
              </Text>
            )}
          </Stack>
          <Flex direction="column" gap="xxs" align={{ base: 'flex-start', md: 'flex-end' }} flex={{ md: 1 }}>
            <Text size="xs" c="gray.6">
              {date}
            </Text>
          </Flex>
        </Flex>
        <Box data-chevron display="flex">
          <Icon name="chevron-right" size={16} />
        </Box>
      </Group>
    </Paper>
  )
}
