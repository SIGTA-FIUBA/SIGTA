import {
  Alert,
  Anchor,
  Box,
  Container,
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
import { useInView } from 'motion/react'
import { useRef, type ReactNode } from 'react'
import { Link, useOutletContext } from 'react-router'
import { useInboxProcedures, useMyProcedures } from '../api/procedures'
import type { components } from '../api/schema'
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

// Figure is one figure of the staff: a large number that counts up to its
// value and what it counts, between the rules of the row of figures. The whole
// cell is a link to the inbox with that filter; under the pointer a tint rises
// from its foot, the number lifts and the arrow moves forward.
function Figure({ value, label, to }: { value: number; label: string; to: string }) {
  return (
    <Paper component={Link} to={to} variant="figure" radius={0} px={{ base: 'md', md: 'lg' }} py="lg">
      <Group justify="space-between" align="flex-end" wrap="nowrap" gap="sm">
        <Stack gap="xs">
          <Text component="div" variant="figure" c="fiuba.9" data-figure>
            <CountUp value={value} />
          </Text>
          <Text size="sm" c="gray.7">
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

// figures are what the staff has in hand: what needs action and what is in
// progress, the two figures the inbox gives.
function figures(list: ProcedureList) {
  return [
    { value: list.items.filter((item) => item.actionable).length, label: 'Requieren acción', to: '/inbox' },
    { value: list.items.length, label: 'En curso', to: '/inbox?view=all' },
  ]
}

// Summary is the row of figures under the greeting, as wide as the page, in
// two cells that close the row. While the list loads, it keeps its place.
function Summary() {
  const procedures = useInboxProcedures()
  if (procedures.isError) {
    return null
  }
  const list = procedures.data
  if (!list) {
    return <Skeleton h="var(--sigta-figures-height)" />
  }
  return (
    <SimpleGrid variant="figures" cols={2} spacing={0} aria-label="Resumen">
      {figures(list).map((figure) => (
        <Figure key={figure.to} {...figure} />
      ))}
    </SimpleGrid>
  )
}

// Condensed takes the place of the greeting and the figures once they leave
// the screen: a thin bar fixed on top with the name and the two figures, so
// the staff keeps them in sight while reading down the page. Hidden, it is out
// of reach of the keyboard and of screen readers, which have the figures
// above.
function Condensed({ shown, greeting }: { shown: boolean; greeting: string }) {
  const procedures = useInboxProcedures()
  const list = procedures.data
  return (
    <Paper variant="condensed" radius={0} data-shown={shown || undefined} aria-hidden={!shown} inert={!shown}>
      <Container px={0}>
        <Group justify="space-between" wrap="nowrap" gap="lg" py="sm">
          <Text fw={600} truncate>
            {greeting}
          </Text>
          {list && (
            <Group gap="lg" wrap="nowrap">
              {figures(list).map((figure) => (
                <Anchor key={figure.to} component={Link} to={figure.to} size="sm" c="gray.7">
                  <Text component="span" fw={700} c="fiuba.9" inherit>
                    {figure.value}
                  </Text>{' '}
                  {figure.label.toLowerCase()}
                </Anchor>
              ))}
            </Group>
          )}
        </Group>
      </Container>
    </Paper>
  )
}

const staffColumns = ['Trámite', 'Paso', 'A cargo de', 'Último movimiento'] as const
const myColumns = ['Trámite', 'Paso', 'Estado', 'Último movimiento'] as const

// standing says who acts now on a procedure of the inbox. One waiting on its
// applicant is in no area's hands, even if its open step belongs to one.
function standing(item: Item) {
  if (item.state === 'with_applicant' || !item.area) {
    return { text: procedureStates[item.state].label }
  }
  return { label: 'A cargo de', text: item.area.name }
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
    <ProcedureRows label="Trámites que requieren acción" columns={staffColumns}>
      {items.slice(0, recentCount).map((item) => (
        <ProcedureRow
          key={item.id}
          to={procedurePath(item.id)}
          state={{ from: '/' }}
          type={item.procedure_type_name}
          applicant={item.applicant.name}
          step={{ name: item.step_name }}
          standing={standing(item)}
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
                <Anchor
                  component={Link}
                  to={procedurePath(item.id)}
                  state={{ from: '/' }}
                  viewTransition
                  c="gray.9"
                  fw={600}
                  inherit
                >
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

// Section is a part of the page on the grid of the page: on a wide screen,
// its title and what it says on the left third and its content on the rest,
// so every section starts on the same line; on a narrow one, the title above.
function Section({
  id,
  title,
  lead,
  action,
  children,
}: {
  id: string
  title: string
  lead?: string
  action?: ReactNode
  children: ReactNode
}) {
  return (
    <Grid component="section" aria-labelledby={id} gap={{ base: 'md', lg: 'xl' }}>
      <Grid.Col span={{ base: 12, lg: 3 }}>
        <Stack gap="xs" align="flex-start">
          <Title order={2} id={id}>
            {title}
          </Title>
          {lead && (
            <Text size="sm" c="gray.7">
              {lead}
            </Text>
          )}
          {action}
        </Stack>
      </Grid.Col>
      <Grid.Col span={{ base: 12, lg: 9 }}>{children}</Grid.Col>
    </Grid>
  )
}

// StaffHome is the work of the staff below the figures: what needs action
// and the latest moves. While the list loads, each part shows its shape.
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
    <>
      <Section
        id="pending"
        title="Requieren acción"
        lead="Los últimos cinco, por su novedad más reciente."
        action={
          <Anchor component={Link} to="/inbox" size="sm" fw={600}>
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
      <Section id="activity" title="Actividad reciente" lead="Dónde quedó cada trámite después de su último paso.">
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
    <ProcedureRows label={label} columns={myColumns}>
      {procedures.map((procedure) => (
        <ProcedureRow
          key={procedure.id}
          to={procedurePath(procedure.id)}
          state={{ from: '/' }}
          type={procedure.procedure_type_name}
          note={procedure.correction_pending ? 'Subsanación pendiente del titular' : undefined}
          step={{ name: procedure.step_name ?? 'Trámite cerrado' }}
          standing={{ text: procedureStates[procedure.state].label }}
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

// HomePage greets the person and shows what to do next, on a white page
// ordered by type, space and rules, without boxes. The staff gets its figures
// under the greeting, which a thin bar keeps in sight once they leave the
// screen, the procedures that need action and the latest moves; everyone, the
// procedures they started, and how SIGTA is used by their profile. The
// destinations live in the menu, so no button here repeats them.
export function HomePage() {
  // The layout passes the person with the session; outside it there is none.
  const user = useOutletContext<User | null | undefined>()
  const staff = user?.staff === true
  const firstName = user?.name.split(' ')[0]
  const greeting = firstName ? `Hola, ${firstName}` : 'Inicio'
  const figuresRef = useRef<HTMLDivElement>(null)
  const figuresInView = useInView(figuresRef, { initial: true })

  return (
    <Stack gap="xxxl">
      <Stack gap="xl">
        <Stack component="header" variant="masthead" gap="xs" aria-labelledby="greeting">
          <Reveal>
            <Text size="sm" c="gray.7">
              {todayText()}
            </Text>
          </Reveal>
          <Reveal index={1}>
            <Title order={1} variant="display" id="greeting" c="fiuba.9">
              {greeting}
            </Title>
          </Reveal>
          {!staff && (
            <Reveal index={2}>
              <Text size="lg" c="gray.7">
                Acá ves en qué paso está cada trámite que iniciaste.
              </Text>
            </Reveal>
          )}
        </Stack>
        {staff && (
          <Box ref={figuresRef}>
            <Reveal index={2}>
              <Summary />
            </Reveal>
          </Box>
        )}
      </Stack>
      {staff && <Condensed shown={!figuresInView} greeting={greeting} />}
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
