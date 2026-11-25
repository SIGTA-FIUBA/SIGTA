import { Alert, Anchor, Box, Button, Grid, Group, List, SimpleGrid, Skeleton, Stack, Text, Title } from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { Link, useOutletContext } from 'react-router'
import { catalogQuery } from '../api/catalog'
import { useApplicantTasks, useInboxProcedures, useMyProcedures } from '../api/procedures'
import type { components } from '../api/schema'
import { EmptyState } from '../components/EmptyState/EmptyState'
import { Icon } from '../components/Icon/Icon'
import { PendingCard } from '../components/PendingCard/PendingCard'
import { PendingRow } from '../components/PendingRow/PendingRow'
import { StartCard } from '../components/StartCard/StartCard'
import { ProcedureRows } from '../components/ProcedureRows/ProcedureRows'
import { Reveal } from '../components/Reveal/Reveal'
import { MyProcedureRows } from '../components/MyProcedureRows/MyProcedureRows'
import { applicantName, byLastMove, movedText, procedurePath } from '../content/procedures'
import { PageHeader } from '../components/PageHeader/PageHeader'

type User = components['schemas']['User']
type ProcedureList = components['schemas']['ProcedureList']

// The home page shows the latest of each list of the staff, no more: the rest
// is in the inbox. An applicant sees every procedure in progress; the closed
// ones are in the history ("Mis trámites"), and the rest of the procedure
// types in the gallery of a new procedure: one row of cards.
const pendingCount = 5
const startCount = 3

// Section is a part of the home page: its title, what it offers on the
// right, and its content below.
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
    <Stack component="section" gap="md" aria-labelledby={id}>
      <Group justify="space-between" align="center" gap="md">
        <Title order={2} id={id}>
          {title}
        </Title>
        {action}
      </Group>
      {children}
    </Stack>
  )
}

function RowsLoading() {
  return (
    <Stack gap="xs" aria-label="Cargando">
      {Array.from({ length: 3 }, (_, index) => (
        <Skeleton key={index} h="var(--sigta-row-height)" />
      ))}
    </Stack>
  )
}

function LoadError() {
  return (
    <Alert color="red" role="alert">
      No se pudieron cargar los trámites. Volvé a intentarlo más tarde.
    </Alert>
  )
}

// ApplicantPending is what the person has to do on their own procedures: each
// correction they have not answered, named by what it asks for, as a card
// that cannot be missed.
function ApplicantPending() {
  const { tasks, isPending, isError } = useApplicantTasks()
  return (
    <Section id="pending" title="Pendientes">
      {isError ? (
        <LoadError />
      ) : isPending ? (
        <RowsLoading />
      ) : tasks.length === 0 ? (
        <EmptyState title="Sin pendientes" text="Cuando un trámite pida algo, aparece acá." />
      ) : (
        <SimpleGrid component="ul" aria-label="Pendientes" cols={1} spacing="sm" p={0} m={0}>
          {tasks.map((task, index) => (
            <Box component="li" key={task.id} display="block">
              <Reveal index={index}>
                <PendingCard
                  to={procedurePath(task.procedureId)}
                  state={{ from: '/' }}
                  task={task.task}
                  context={task.context}
                />
              </Reveal>
            </Box>
          ))}
        </SimpleGrid>
      )}
    </Section>
  )
}

// NewProcedure offers the first types of the catalog, on top of the person's
// procedures: a card each, as in the gallery of a new procedure, and the way
// to all of them. The search and the start are on top of every page too
// (AppLayout).
function NewProcedure() {
  const catalog = useQuery(catalogQuery)
  const entries = catalog.data?.slice(0, startCount) ?? []
  return (
    <Section
      id="new"
      title="Nuevo trámite"
      action={
        <Button
          component={Link}
          to="/new"
          variant="subtle"
          size="compact-sm"
          rightSection={<Icon name="arrow-right" size={16} />}
        >
          Ver todos
        </Button>
      }
    >
      {catalog.isError ? (
        <Alert color="red" role="alert">
          No se pudo cargar el catálogo. Volvé a intentarlo más tarde.
        </Alert>
      ) : catalog.isPending ? (
        <RowsLoading />
      ) : entries.length === 0 ? (
        <EmptyState title="Sin trámites publicados" text="Cuando se publique un tipo de trámite, aparece acá." />
      ) : (
        <SimpleGrid
          component="ul"
          aria-label="Trámites para iniciar"
          cols={{ base: 1, md: startCount }}
          spacing={{ base: 'xs', sm: 'md' }}
          p={0}
          m={0}
        >
          {entries.map((entry, index) => (
            <Box component="li" key={entry.id} display="block">
              <Reveal index={index}>
                <StartCard to={`/catalog/${entry.id}/start`} name={entry.name} description={entry.description} />
              </Reveal>
            </Box>
          ))}
        </SimpleGrid>
      )}
    </Section>
  )
}

// InProgress is every procedure of the person that is not closed, by its
// last move; the closed ones are in the history.
function InProgress() {
  const myProcedures = useMyProcedures()
  const inProgress = (myProcedures.data ?? []).filter((procedure) => procedure.state !== 'closed').sort(byLastMove)
  return (
    <Section
      id="in-progress"
      title="Trámites en curso"
      action={
        <Button
          component={Link}
          to="/my-procedures"
          variant="subtle"
          size="compact-sm"
          rightSection={<Icon name="arrow-right" size={16} />}
        >
          Historial
        </Button>
      }
    >
      {myProcedures.isError ? (
        <LoadError />
      ) : myProcedures.isPending ? (
        <RowsLoading />
      ) : inProgress.length === 0 ? (
        <EmptyState title="Sin trámites en curso" text="Los trámites que se inicien aparecen acá." />
      ) : (
        <MyProcedureRows label="Trámites en curso" procedures={inProgress} from="/" />
      )}
    </Section>
  )
}

// StaffPending is what needs action from the staff, the latest five: on top
// the step to resolve, below the procedure and its applicant.
function StaffPending({ list }: { list: ProcedureList }) {
  const items = list.items.filter((item) => item.actionable).sort(byLastMove)
  return (
    <Section
      id="pending"
      title="Pendientes"
      action={
        items.length > pendingCount && (
          <Button
            component={Link}
            to="/inbox"
            variant="subtle"
            size="compact-sm"
            rightSection={<Icon name="arrow-right" size={16} />}
          >
            Ver todos
          </Button>
        )
      }
    >
      {items.length === 0 ? (
        <EmptyState title="Sin pendientes" text="Cuando un trámite requiera acción, aparece acá." />
      ) : (
        <ProcedureRows label="Pendientes">
          {items.slice(0, pendingCount).map((item) => (
            <PendingRow
              key={item.id}
              to={procedurePath(item.id)}
              state={{ from: '/' }}
              task={item.step_name}
              context={`${item.procedure_type_name}, ${applicantName(item)}`}
            />
          ))}
        </ProcedureRows>
      )}
    </Section>
  )
}

// Activity lists the latest moves of every visible procedure, one entry each,
// with the step where it is now. The entries hang from a line, the newest on
// top.
function Activity({ list }: { list: ProcedureList }) {
  const items = [...list.items].sort(byLastMove).slice(0, pendingCount)
  return (
    <Section id="activity" title="Actividad reciente">
      {items.length === 0 ? (
        <Text c="dimmed">Todavía no hay movimientos.</Text>
      ) : (
        <Reveal>
          <List variant="timeline" size="sm" aria-label="Movimientos recientes">
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
                      {item.procedure_type_name} de {applicantName(item)}
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
      )}
    </Section>
  )
}

// StaffHome is the work of the staff: what needs action and, beside it from
// breakpoint.xl, the latest moves; narrower, the moves go below, since beside
// the pending ones they would wrap every few words.
function StaffHome() {
  const procedures = useInboxProcedures()
  if (procedures.isError) {
    return <LoadError />
  }
  const list = procedures.data
  if (!list) {
    return <RowsLoading />
  }
  return (
    <Grid gap="xxl">
      <Grid.Col span={{ base: 12, xl: 8 }}>
        <StaffPending list={list} />
      </Grid.Col>
      <Grid.Col span={{ base: 12, xl: 4 }}>
        <Activity list={list} />
      </Grid.Col>
    </Grid>
  )
}

// HomePage is what to do next, without a greeting: for the staff, what needs
// action and the latest moves; for an applicant, the procedures to start and
// their own: what they ask for and those in progress.
export function HomePage() {
  // The layout passes the person with the session; outside it there is none.
  const user = useOutletContext<User | null | undefined>()
  const staff = user?.staff === true

  return (
    <Stack gap="xl">
      {/* An applicant starts a procedure from the top right of every page
          (AppLayout). */}
      <PageHeader title={staff ? 'Inicio' : 'Mis trámites'} />
      {staff ? (
        <StaffHome />
      ) : (
        // One column, in the order of what the person comes for: start a
        // procedure, answer what one asks for, follow the rest.
        <Stack gap="xxl">
          <NewProcedure />
          <ApplicantPending />
          <InProgress />
        </Stack>
      )}
    </Stack>
  )
}
