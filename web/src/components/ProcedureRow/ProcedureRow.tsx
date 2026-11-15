import { Box, Flex, Group, Paper, Stack, Text } from '@mantine/core'
import { Children, type ReactNode } from 'react'
import { Link } from 'react-router'
import { Icon } from '../Icon/Icon'
import { NewsDot } from '../NewsDot/NewsDot'
import { StatusBadge, type StatusTone } from '../StatusBadge/StatusBadge'
import { StepBar } from '../StepBar/StepBar'

export type ProcedureRowProps = {
  to: string
  // state goes with the link, such as the page to come back to.
  state?: unknown
  // type is the name of the procedure type; fileNumber, the electronic file
  // once the staff registered it.
  type: string
  fileNumber?: string
  applicant: string
  program?: string
  // step is the open step and where it sits among the steps of the type.
  step: { number: number; count: number; name: string }
  // area is who acts now.
  area: string
  status: { tone: StatusTone; label: string }
  // date is the last news, already written for people; due, a deadline if
  // there is one.
  date: string
  due?: string
  // news marks a procedure that moved since the person last saw it.
  news?: boolean
}

// ProcedureRow is one procedure in a list, in four zones that answer what it
// is, whose it is, at which stage it is and who acts now, with its state and
// its last news. The whole row is one link.
export function ProcedureRow({
  to,
  state,
  type,
  fileNumber,
  applicant,
  program,
  step,
  area,
  status,
  date,
  due,
  news,
}: ProcedureRowProps) {
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
            <Group gap="xs" wrap="nowrap" miw={0}>
              {news && <NewsDot />}
              <Text fw={600} truncate>
                {type}
              </Text>
            </Group>
            {fileNumber && (
              <Text size="xs" c="gray.6">
                {fileNumber}
              </Text>
            )}
            <Text size="sm" c="gray.7">
              {program ? `${applicant}, ${program}` : applicant}
            </Text>
          </Stack>
          <Stack gap="xs" miw={0} flex={{ md: 2 }}>
            <StepBar step={step.number} count={step.count} />
            <Text size="xs" c="gray.7">
              {step.name}
            </Text>
          </Stack>
          <Stack gap="xxs" align="flex-start" miw={0} flex={{ md: 2 }}>
            <StatusBadge tone={status.tone}>{status.label}</StatusBadge>
            <Text size="xs" c="gray.7">
              A cargo de {area}
            </Text>
          </Stack>
          <Flex direction="column" gap="xxs" align={{ base: 'flex-start', md: 'flex-end' }} flex={{ md: 1 }}>
            <Text size="xs" c="gray.6">
              {date}
            </Text>
            {due && (
              <Text size="xs" c="orange.7">
                Vence {due}
              </Text>
            )}
          </Flex>
        </Flex>
        <Box data-chevron display="flex">
          <Icon name="chevron-right" size={16} />
        </Box>
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
