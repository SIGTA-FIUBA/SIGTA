import { ActionIcon, Anchor, Divider, Indicator, Popover, Stack, Text } from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'
import { Link } from 'react-router'
import { useApplicantTasks, useMyProcedures } from '../api/procedures'
import { Icon } from '../components/Icon/Icon'
import { byLastMove, movedText, procedurePath, stateText } from '../content/procedures'

// The bell shows the latest moves, no more: the rest is in the history.
const movesCount = 5

// NotificationsBell is what changed on the person's own procedures, from what
// the pages already ask for: what they have to do, marked with a dot on the
// bell, and the latest move of each procedure in progress. Read and unread
// notifications would need the server, so there are none.
export function NotificationsBell() {
  const [opened, { toggle, close }] = useDisclosure(false)
  const { tasks } = useApplicantTasks()
  const mine = useMyProcedures()
  const moves = (mine.data ?? [])
    .filter((procedure) => procedure.state !== 'closed')
    .sort(byLastMove)
    .slice(0, movesCount)
  const label = tasks.length > 0 ? 'Notificaciones, con acción requerida' : 'Notificaciones'

  return (
    <Popover opened={opened} onChange={(next) => !next && close()} position="bottom-end" width={320}>
      <Popover.Target>
        <Indicator disabled={tasks.length === 0} color="attention.6" size={8} offset={6}>
          <ActionIcon variant="subtle" color="gray" size="lg" aria-label={label} onClick={toggle}>
            <Icon name="bell" />
          </ActionIcon>
        </Indicator>
      </Popover.Target>
      <Popover.Dropdown p="md">
        <Stack gap="md">
          {tasks.length > 0 && (
            <Stack gap="xs">
              <Text size="sm" fw={600}>
                Acción requerida
              </Text>
              {tasks.map((task) => (
                <Stack key={task.id} gap={0}>
                  <Anchor component={Link} to={procedurePath(task.procedureId)} size="sm" fw={600} onClick={close}>
                    {task.task}
                  </Anchor>
                  <Text size="sm" c="gray.7">
                    {task.context}
                  </Text>
                </Stack>
              ))}
            </Stack>
          )}
          {tasks.length > 0 && moves.length > 0 && <Divider />}
          {moves.length > 0 ? (
            <Stack gap="xs">
              <Text size="sm" fw={600}>
                Últimos movimientos
              </Text>
              {moves.map((procedure) => (
                <Stack key={procedure.id} gap={0}>
                  <Anchor
                    component={Link}
                    to={procedurePath(procedure.id)}
                    size="sm"
                    c="gray.9"
                    fw={600}
                    onClick={close}
                  >
                    {procedure.procedure_type_name}
                  </Anchor>
                  <Text size="sm" c="gray.7">
                    {procedure.step_name}, {stateText(procedure.state, { area: procedure.area, applicant: true }).text}
                  </Text>
                  <Text size="xs" c="gray.6">
                    {movedText(procedure.last_moved_at)}
                  </Text>
                </Stack>
              ))}
            </Stack>
          ) : (
            tasks.length === 0 && (
              <Text size="sm" c="gray.7">
                Sin novedades
              </Text>
            )
          )}
        </Stack>
      </Popover.Dropdown>
    </Popover>
  )
}
