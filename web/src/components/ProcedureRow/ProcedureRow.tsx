import { Box, Group, Paper, Stack, Text } from '@mantine/core'
import { Link, useViewTransitionState } from 'react-router'
import { Icon } from '../Icon/Icon'
import { rowColumns } from './columns'

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
  // standing is the third column: who acts now in the lists of the staff
  // ("A cargo de" an area), the state in the lists of the person's own
  // procedures. Its label is said on a phone and by the column heading on a
  // wide screen.
  standing: { label?: string; text: string }
  // note is something the person has to do, such as a correction.
  note?: string
  // date is the last news, already written for people.
  date: string
  // age is how long the procedure has been in progress, under the date, for
  // the lists that sort by it.
  age?: string
}

// ProcedureRow is one procedure in a list, in four columns that answer what
// it is and whose it is, at which stage it is, who acts now or in which state
// it is, and its last news. The whole row is one link, with no box: the list
// draws a rule between the rows and a mark that follows the pointer. Opening
// it, the type becomes the title of the page it opens.
export function ProcedureRow({ to, state, type, applicant, step, standing, note, date, age }: ProcedureRowProps) {
  const opening = useViewTransitionState(to)
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
      <Group wrap="nowrap" gap="md">
        {/* On a phone, what it is with its date at the right, then where it
            is and who has it, side by side. From md on the four columns share
            the row (interactions.module.css). */}
        <Box data-cells>
          <Stack data-cell="what" gap="xxs" miw={0} flex={{ md: rowColumns[0] }}>
            {/* Cut from md on, the name stays readable on hover. */}
            <Text fw={600} title={type} data-transition-title={opening || undefined}>
              {type}
            </Text>
            {applicant && (
              <Text size="sm" c="gray.7">
                {applicant}
              </Text>
            )}
            {note && (
              <Text size="sm" fw={600} c="orange.8">
                {note}
              </Text>
            )}
          </Stack>
          <Box data-cell="where">
            <Text size="sm" miw={0} flex={{ md: rowColumns[1] }}>
              {step.name}
            </Text>
            <Text size="sm" c="gray.7" miw={0} flex={{ md: rowColumns[2] }}>
              {standing.label && (
                <Text component="span" variant="prefix" inherit>
                  {standing.label}{' '}
                </Text>
              )}
              {standing.text}
            </Text>
          </Box>
          <Stack data-cell="when" gap="xxs" ta="right" flex={{ md: rowColumns[3] }}>
            <Text size="sm" c="gray.7" data-cell="date">
              {date}
            </Text>
            {age && (
              <Text size="xs" c="gray.7" data-cell="age">
                {age}
              </Text>
            )}
          </Stack>
        </Box>
        <Box data-chevron aria-hidden>
          <Icon name="chevron-right" size={16} />
        </Box>
      </Group>
    </Paper>
  )
}
