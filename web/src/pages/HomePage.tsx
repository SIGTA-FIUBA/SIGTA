import { Alert, Anchor, Box, Group, List, Paper, Skeleton, Stack, Text, Title } from '@mantine/core'
import type { ReactNode } from 'react'
import { Link, useOutletContext } from 'react-router'
import { useInboxProcedures, useMyProcedures } from '../api/procedures'
import type { components } from '../api/schema'
import { EmptyState } from '../components/EmptyState/EmptyState'
import { ProcedureRow } from '../components/ProcedureRow/ProcedureRow'
import { ProcedureRows } from '../components/ProcedureRows/ProcedureRows'
import { StepList } from '../components/StepList/StepList'
import { applicantHowTo } from '../content/platform'
import { movedText, procedureStates } from '../content/procedures'

type User = components['schemas']['User']
type Item = components['schemas']['ProcedureListItem']
type ProcedureList = components['schemas']['ProcedureList']
type MyProcedure = components['schemas']['MyProcedure']

// The home page shows the latest procedures that need action and the latest
// news, no more: the rest is in the inbox.
const latest = 5

const dateFormat = new Intl.DateTimeFormat('es-AR', { weekday: 'long', day: 'numeric', month: 'long' })

// todayText writes today's date for the greeting: "Sábado, 14 de noviembre".
function todayText() {
  const text = dateFormat.format(new Date())
  return text.charAt(0).toUpperCase() + text.slice(1)
}

function byLastMove(a: Item, b: Item) {
  return Date.parse(b.last_moved_at) - Date.parse(a.last_moved_at)
}

function procedureLink(item: Item) {
  return `/procedures/${encodeURIComponent(item.id)}`
}

// Figure is one cell of the summary: a number and what it counts. With a
// destination, the whole cell is a link to the inbox with that filter.
function Figure({ value, label, to }: { value: number; label: string; to?: string }) {
  const content = (
    <Stack gap="xxs">
      <Text data-figure fz="xl" lh="xl" fw={600}>
        {value}
      </Text>
      <Text size="xs" c="gray.7">
        {label}
      </Text>
    </Stack>
  )
  if (to === undefined) {
    return <Box p="md">{content}</Box>
  }
  return (
    <Paper component={Link} to={to} variant="interactive" radius={0} p="md" td="none" c="inherit">
      {content}
    </Paper>
  )
}

function Summary({ list }: { list: ProcedureList }) {
  const actionable = list.items.filter((item) => item.actionable).length
  return (
    <Paper withBorder variant="figures" p={0} aria-label="Resumen">
      <Figure value={actionable} label="Requieren acción" to="/inbox" />
      <Figure value={list.items.length} label="En curso" to="/inbox?view=all" />
    </Paper>
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
      {items.slice(0, latest).map((item) => (
        <ProcedureRow
          key={item.id}
          to={procedureLink(item)}
          state={{ from: '/' }}
          type={item.procedure_type_name}
          applicant={item.applicant.name}
          step={{ name: item.step_name }}
          area={item.area?.name}
          status={procedureStates[item.state]}
          date={movedText(item.last_moved_at)}
        />
      ))}
    </ProcedureRows>
  )
}

// Activity lists the latest moves of every visible procedure, one line each.
function Activity({ list }: { list: ProcedureList }) {
  const items = [...list.items].sort(byLastMove).slice(0, latest)
  if (items.length === 0) {
    return <Text c="dimmed">Todavía no hay movimientos.</Text>
  }
  return (
    <List listStyleType="none" spacing="sm" aria-label="Movimientos recientes">
      {items.map((item) => (
        <List.Item key={item.id}>
          <Group justify="space-between" wrap="nowrap" gap="md">
            <Text size="sm" truncate>
              <Anchor component={Link} to={procedureLink(item)} state={{ from: '/' }} inherit>
                {item.procedure_type_name} de {item.applicant.name}
              </Anchor>{' '}
              pasó a {item.step_name}
            </Text>
            <Text size="xs" c="gray.6" flex="none">
              {movedText(item.last_moved_at)}
            </Text>
          </Group>
        </List.Item>
      ))}
    </List>
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

// StaffHome is the home of the staff: the summary, what needs action and the
// latest moves. While the list loads, each part shows its shape.
function StaffHome() {
  const procedures = useInboxProcedures()
  if (procedures.isError) {
    return <EmptyState title="No se pudieron cargar los trámites" text="Volvé a intentarlo en unos minutos." />
  }
  const list = procedures.data
  return (
    <>
      {list ? <Summary list={list} /> : <Skeleton h="var(--sigta-figures-height)" />}
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
      <Section id="activity" title="Actividad reciente">
        {list ? <Activity list={list} /> : <Skeleton h="var(--sigta-row-height)" />}
      </Section>
    </>
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
          to={`/procedures/${encodeURIComponent(procedure.id)}`}
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

// HomePage greets the person and shows what to do next. The staff gets the
// summary, the procedures that need action and the latest moves; everyone,
// the procedures they started; the applicant, how a procedure is done. The destinations live in the menu, so
// no button here repeats them.
export function HomePage() {
  // The layout passes the person with the session; outside it there is none.
  const user = useOutletContext<User | null | undefined>()
  const staff = user?.staff === true
  const firstName = user?.name.split(' ')[0]

  return (
    <Stack gap="xxl">
      <Stack gap="xxs">
        <Title order={1}>{firstName ? `Hola, ${firstName}` : 'Inicio'}</Title>
        <Text size="sm" c="gray.7">
          {todayText()}
        </Text>
      </Stack>
      {staff && <StaffHome />}
      <MyProcedures />
      {!staff && (
        <Section id="how" title="Cómo se realiza un trámite">
          <StepList steps={applicantHowTo} />
        </Section>
      )}
    </Stack>
  )
}
