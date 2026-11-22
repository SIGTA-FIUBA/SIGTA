import {
  Alert,
  Anchor,
  Box,
  Flex,
  Grid,
  Group,
  List,
  Paper,
  SimpleGrid,
  Skeleton,
  Stack,
  Text,
  Title,
} from '@mantine/core'
import type { ReactNode } from 'react'
import { Link, useOutletContext } from 'react-router'
import { useInboxProcedures, useMyProcedures } from '../api/procedures'
import type { components } from '../api/schema'
import { BlueprintPanel } from '../components/BlueprintPanel/BlueprintPanel'
import { CountUp } from '../components/CountUp/CountUp'
import { EmptyState } from '../components/EmptyState/EmptyState'
import { Icon } from '../components/Icon/Icon'
import { ProcedureRow } from '../components/ProcedureRow/ProcedureRow'
import { ProcedureRows } from '../components/ProcedureRows/ProcedureRows'
import { Reveal } from '../components/Reveal/Reveal'
import { StepList } from '../components/StepList/StepList'
import { applicantHowTo, staffHowTo } from '../content/platform'
import { movedText, procedurePath, procedureStates } from '../content/procedures'

type User = components['schemas']['User']
type Item = components['schemas']['ProcedureListItem']
type ProcedureList = components['schemas']['ProcedureList']
type MyProcedure = components['schemas']['MyProcedure']

// The home page shows the latest procedures that need action and the latest
// news, no more: the rest is in the inbox.
const recentCount = 5

const dateFormat = new Intl.DateTimeFormat('es-AR', { weekday: 'long', day: 'numeric', month: 'long' })

// todayText writes today's date for the greeting: "Sábado, 14 de noviembre".
function todayText() {
  const text = dateFormat.format(new Date())
  return text.charAt(0).toUpperCase() + text.slice(1)
}

function byLastMove(a: Item, b: Item) {
  return Date.parse(b.last_moved_at) - Date.parse(a.last_moved_at)
}

// Figure is one cell of the summary in the header: a number that counts up
// to its value and what it counts. The whole cell is a link to the inbox with
// that filter, and its arrow moves forward under the pointer.
function Figure({ value, label, to }: { value: number; label: string; to: string }) {
  return (
    <Paper component={Link} to={to} variant="on-blueprint" radius="sm" p="md">
      <Group justify="space-between" align="flex-end" wrap="nowrap" gap="sm">
        <Stack gap="xs">
          <Text component="div" variant="figure">
            <CountUp value={value} />
          </Text>
          <Text size="sm" c="fiuba.1">
            {label}
          </Text>
        </Stack>
        <Box data-chevron aria-hidden>
          <Icon name="chevron-right" size={16} />
        </Box>
      </Group>
    </Paper>
  )
}

// Summary is what the staff has in hand, in the header: what needs action and
// what is in progress, the two figures the inbox gives. While the list loads,
// it keeps its place.
function Summary() {
  const procedures = useInboxProcedures()
  if (procedures.isError) {
    return null
  }
  const list = procedures.data
  if (!list) {
    return <Skeleton h="var(--sigta-figures-height)" />
  }
  const actionable = list.items.filter((item) => item.actionable).length
  return (
    <SimpleGrid cols={2} spacing="sm" aria-label="Resumen">
      <Figure value={actionable} label="Requieren acción" to="/inbox" />
      <Figure value={list.items.length} label="En curso" to="/inbox?view=all" />
    </SimpleGrid>
  )
}

function Pending({ list }: { list: ProcedureList }) {
  const items = list.items.filter((item) => item.actionable).sort(byLastMove)
  if (items.length === 0) {
    return (
      <EmptyState
        title="No hay trámites pendientes"
        text="Cuando un trámite requiera acción de la cuenta, aparece acá."
      />
    )
  }
  return (
    <ProcedureRows label="Trámites que requieren acción">
      {items.slice(0, recentCount).map((item) => (
        <ProcedureRow
          key={item.id}
          to={procedurePath(item.id)}
          state={{ from: '/' }}
          type={item.procedure_type_name}
          applicant={item.applicant.name}
          step={{ name: item.step_name }}
          // A procedure waiting on its applicant is in no area's hands, even
          // if its open step belongs to one.
          area={item.state === 'with_applicant' ? undefined : item.area?.name}
          status={procedureStates[item.state]}
          date={movedText(item.last_moved_at)}
        />
      ))}
    </ProcedureRows>
  )
}

// Activity lists the latest moves of every visible procedure, one entry each,
// with the step where it is now: the list does not say what the move was. The
// entries hang from a line, the newest on top. A long one wraps instead of
// running past the screen of a phone.
function Activity({ list }: { list: ProcedureList }) {
  const items = [...list.items].sort(byLastMove).slice(0, recentCount)
  if (items.length === 0) {
    return <Text c="dimmed">Todavía no hay movimientos.</Text>
  }
  return (
    <Reveal>
      <List variant="timeline" aria-label="Movimientos recientes">
        {items.map((item) => (
          <List.Item key={item.id}>
            <Stack gap="xxs">
              <Text size="sm">
                <Anchor component={Link} to={procedurePath(item.id)} state={{ from: '/' }} inherit>
                  {item.procedure_type_name} de {item.applicant.name}
                </Anchor>{' '}
                está en {item.step_name}
              </Text>
              <Text size="xs" c="gray.6">
                {movedText(item.last_moved_at)}
              </Text>
            </Stack>
          </List.Item>
        ))}
      </List>
    </Reveal>
  )
}

function Section({
  id,
  title,
  action,
  children,
}: {
  id: string
  title: string
  action?: ReactNode
  children: ReactNode
}) {
  return (
    <Stack component="section" aria-labelledby={id} gap="md">
      <Group justify="space-between" align="baseline">
        <Title order={2} id={id}>
          {title}
        </Title>
        {action}
      </Group>
      {children}
    </Stack>
  )
}

// StaffHome is the work of the staff below the header: what needs action,
// and beside it on a wide screen, the latest moves. While the list loads, each
// part shows its shape.
function StaffHome() {
  const procedures = useInboxProcedures()
  if (procedures.isError) {
    return (
      <Alert color="red" role="alert">
        No se pudieron cargar los trámites. Volvé a intentarlo más tarde.
      </Alert>
    )
  }
  const list = procedures.data
  return (
    <Grid gap="xxl">
      <Grid.Col span={{ base: 12, lg: 8 }}>
        <Section
          id="pending"
          title="Requieren acción"
          action={
            <Anchor component={Link} to="/inbox" size="sm">
              Ver todos
            </Anchor>
          }
        >
          {list ? (
            <Pending list={list} />
          ) : (
            <Stack gap="xs" aria-label="Cargando los trámites">
              {Array.from({ length: 3 }, (_, index) => (
                <Skeleton key={index} h="var(--sigta-row-height)" />
              ))}
            </Stack>
          )}
        </Section>
      </Grid.Col>
      <Grid.Col span={{ base: 12, lg: 4 }}>
        <Section id="activity" title="Actividad reciente">
          {list ? <Activity list={list} /> : <Skeleton h="var(--sigta-row-height)" />}
        </Section>
      </Grid.Col>
    </Grid>
  )
}

const closeFormat = new Intl.DateTimeFormat('es-AR', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  timeZone: 'America/Argentina/Buenos_Aires',
})

function MyProcedureRows({ label, procedures }: { label: string; procedures: MyProcedure[] }) {
  return (
    <ProcedureRows label={label}>
      {procedures.map((procedure) => (
        <ProcedureRow
          key={procedure.id}
          to={procedurePath(procedure.id)}
          state={{ from: '/' }}
          type={procedure.procedure_type_name}
          note={procedure.correction_pending ? 'Subsanación pendiente del titular' : undefined}
          step={{ name: procedure.step_name ?? 'Trámite cerrado' }}
          status={procedureStates[procedure.state]}
          date={
            procedure.state === 'closed'
              ? `Cierre: ${closeFormat.format(new Date(procedure.last_moved_at))}`
              : movedText(procedure.last_moved_at)
          }
        />
      ))}
    </ProcedureRows>
  )
}

// MyProcedures lists what the person started, staff or not, so that signing
// in answers where each procedure is. The closed ones go apart, for their
// applicant to reach the resolution.
function MyProcedures() {
  const myProcedures = useMyProcedures()
  const inProgress = myProcedures.data?.filter((procedure) => procedure.state !== 'closed') ?? []
  const closed = myProcedures.data?.filter((procedure) => procedure.state === 'closed') ?? []
  return (
    <>
      <Section id="my-procedures" title="Mis trámites en curso">
        {myProcedures.isPending && (
          <Stack gap="xs" aria-label="Cargando los trámites">
            <Skeleton h="var(--sigta-row-height)" />
          </Stack>
        )}
        {myProcedures.isError && (
          <Alert color="red" role="alert">
            No se pudieron cargar los trámites. Volvé a intentarlo más tarde.
          </Alert>
        )}
        {myProcedures.isSuccess && inProgress.length === 0 && (
          <Text c="dimmed">No hay trámites en curso. Los trámites se inician desde el catálogo.</Text>
        )}
        {inProgress.length > 0 && <MyProcedureRows label="Mis trámites en curso" procedures={inProgress} />}
      </Section>
      {closed.length > 0 && (
        <Section id="my-closed-procedures" title="Mis trámites cerrados">
          <MyProcedureRows label="Mis trámites cerrados" procedures={closed} />
        </Section>
      )}
    </>
  )
}

// HomePage greets the person in a navy header with the grid of a blueprint,
// and shows what to do next. The staff gets its figures in the header, the
// procedures that need action and the latest moves; everyone, the procedures
// they started, and how SIGTA is used by their profile. The destinations live
// in the menu, so no button here repeats them.
export function HomePage() {
  // The layout passes the person with the session; outside it there is none.
  const user = useOutletContext<User | null | undefined>()
  const staff = user?.staff === true
  const firstName = user?.name.split(' ')[0]

  return (
    <Stack gap="xxl">
      <BlueprintPanel component="header" radius="md" p={{ base: 'lg', md: 'xl' }} aria-labelledby="greeting">
        <Flex
          direction={{ base: 'column', md: 'row' }}
          justify="space-between"
          align={{ base: 'stretch', md: 'flex-end' }}
          gap="xl"
        >
          <Reveal>
            <Stack gap="xs">
              <Text size="sm" c="fiuba.1">
                {todayText()}
              </Text>
              <Title order={1} variant="display" id="greeting" c="white">
                {firstName ? `Hola, ${firstName}` : 'Inicio'}
              </Title>
              {!staff && <Text c="fiuba.1">Acá ves en qué paso está cada trámite que iniciaste.</Text>}
            </Stack>
          </Reveal>
          {staff && (
            <Reveal index={2}>
              <Summary />
            </Reveal>
          )}
        </Flex>
      </BlueprintPanel>
      {staff && <StaffHome />}
      <MyProcedures />
      {staff ? (
        <Section id="how" title="Cómo se trabaja en SIGTA">
          <StepList steps={staffHowTo} />
        </Section>
      ) : (
        <Section id="how" title="Cómo se realiza un trámite">
          <StepList steps={applicantHowTo} />
        </Section>
      )}
    </Stack>
  )
}
