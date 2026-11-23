import { Alert, Anchor, Box, Container, Group, List, Paper, Skeleton, Stack, Text, Title } from '@mantine/core'
import { useInView } from 'motion/react'
import { useRef } from 'react'
import { Link, useOutletContext } from 'react-router'
import { useInboxProcedures, useMyProcedures } from '../api/procedures'
import type { components } from '../api/schema'
import { Figure } from '../components/Figure/Figure'
import { Figures } from '../components/Figures/Figures'
import { DocumentFile } from '../components/DocumentFile/DocumentFile'
import { EmptyState } from '../components/EmptyState/EmptyState'
import { ProcedureRow } from '../components/ProcedureRow/ProcedureRow'
import { ProcedureRows } from '../components/ProcedureRows/ProcedureRows'
import { Reveal } from '../components/Reveal/Reveal'
import { PageSection } from '../components/PageSection/PageSection'
import { StepList } from '../components/StepList/StepList'
import { applicantHowTo, staffHowTo } from '../content/platform'
import { closedWithoutResolution, movedText, procedurePath, procedureStates, standing } from '../content/procedures'

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
    <Figures label="Resumen">
      {figures(list).map((figure) => (
        <Figure key={figure.to} {...figure} />
      ))}
    </Figures>
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
      <PageSection
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
      </PageSection>
      <PageSection id="activity" title="Actividad reciente" lead="Dónde quedó cada trámite después de su último paso.">
        {list ? <Activity list={list} /> : <Skeleton h="var(--sigta-row-height)" />}
      </PageSection>
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

// ClosedResolutions goes under the rows of the closed procedures, apart from
// them because each row is a link as a whole: for each one, the resolution to
// see or download, or that it closed without one.
function ClosedResolutions({ procedures }: { procedures: MyProcedure[] }) {
  return (
    <Stack gap="lg">
      {procedures.map((procedure) => (
        <Stack key={procedure.id} gap="xs">
          <Text fw={600}>{procedure.procedure_type_name}</Text>
          {procedure.resolutions.length === 0 ? (
            <Text size="sm" c="gray.7">
              {closedWithoutResolution}
            </Text>
          ) : (
            procedure.resolutions.map((resolution) => (
              <Stack key={resolution.id} gap="xxs">
                <Text size="sm" fw={500}>
                  {resolution.document_name}
                </Text>
                <DocumentFile
                  procedureId={procedure.id}
                  file={resolution}
                  procedureName={procedure.procedure_type_name}
                />
              </Stack>
            ))
          )}
        </Stack>
      ))}
    </Stack>
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
      <PageSection id="my-procedures" title="Mis trámites en curso">
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
      </PageSection>
      {closed.length > 0 && (
        <PageSection id="my-closed-procedures" title="Mis trámites cerrados">
          <Stack gap="xl">
            <MyProcedureRows label="Mis trámites cerrados" procedures={closed} />
            <ClosedResolutions procedures={closed} />
          </Stack>
        </PageSection>
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
        <PageSection id="how" title="Cómo se trabaja en SIGTA">
          <StepList steps={staffHowTo} />
        </PageSection>
      ) : (
        <PageSection id="how" title="Cómo se realiza un trámite">
          <StepList steps={applicantHowTo} />
        </PageSection>
      )}
    </Stack>
  )
}
