import { Alert, Anchor, Badge, Group, Loader, Paper, SimpleGrid, Stack, Text, ThemeIcon, Title } from '@mantine/core'
import { Link, useOutletContext } from 'react-router'
import { useMyProcedures } from '../api/procedures'
import type { components } from '../api/schema'
import { DocumentFile } from '../components/DocumentFile/DocumentFile'
import { applicantHowTo, principles, staffHowTo } from '../content/platform'
import { closedWithoutResolution, stateColors, stateNames } from '../content/procedureStates'

type User = components['schemas']['User']
type MyProcedure = components['schemas']['MyProcedure']

const dateFormat = new Intl.DateTimeFormat('es-AR', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  timeZone: 'America/Argentina/Buenos_Aires',
})

function MyProcedureCard({ procedure }: { procedure: MyProcedure }) {
  return (
    <Paper
      component={Link}
      to={`/procedures/${encodeURIComponent(procedure.id)}`}
      variant="interactive"
      withBorder
      p="md"
      h="fit-content"
      td="none"
      c="inherit"
    >
      <Stack gap="xs">
        <Group justify="space-between" gap="xs" align="flex-start">
          <Text fw={600} c="fiuba.7">
            {procedure.procedure_type_name}
          </Text>
          <Badge variant="light" color={stateColors[procedure.state]}>
            {stateNames[procedure.state]}
          </Badge>
        </Group>
        {procedure.correction_pending && (
          <Text size="sm" fw={600} c="fiuba.7">
            Subsanación pendiente del titular
          </Text>
        )}
        <Text size="sm">Paso actual: {procedure.step_name}</Text>
        <Text size="sm" c="dimmed">
          Último movimiento: {dateFormat.format(new Date(procedure.last_moved_at))}
        </Text>
      </Stack>
    </Paper>
  )
}

// ClosedProcedureCard is not a link as a whole, as MyProcedureCard is, because
// it holds the buttons to see and download the resolution: its name leads to
// the detail. It has no state badge, which would repeat the title of its
// section.
function ClosedProcedureCard({ procedure }: { procedure: MyProcedure }) {
  return (
    <Paper withBorder p="md" h="fit-content">
      <Stack gap="xs">
        <Anchor component={Link} to={`/procedures/${encodeURIComponent(procedure.id)}`} fw={600}>
          {procedure.procedure_type_name}
        </Anchor>
        <Text size="sm" c="dimmed">
          Cierre: {dateFormat.format(new Date(procedure.last_moved_at))}
        </Text>
        {procedure.resolutions.length === 0 ? (
          <Text size="sm" c="dimmed">
            {closedWithoutResolution}
          </Text>
        ) : (
          procedure.resolutions.map((resolution) => (
            <Stack key={resolution.id} gap={0}>
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
    </Paper>
  )
}

// MyProcedureGrid keeps each card as tall as its content (h="fit-content" on
// the cards), so that a short card does not stretch to the tallest one of its
// row.
function MyProcedureGrid({ procedures }: { procedures: MyProcedure[] }) {
  return (
    <SimpleGrid cols={{ base: 1, sm: 2, md: 3 }} spacing="sm">
      {procedures.map((procedure) =>
        procedure.state === 'closed' ? (
          <ClosedProcedureCard key={procedure.id} procedure={procedure} />
        ) : (
          <MyProcedureCard key={procedure.id} procedure={procedure} />
        ),
      )}
    </SimpleGrid>
  )
}

// MyProceduresSection lists what the person started, so that signing in
// answers where each procedure is. The closed ones go apart, with their
// resolution.
function MyProceduresSection() {
  const myProcedures = useMyProcedures()
  const inProgress = myProcedures.data?.filter((procedure) => procedure.state !== 'closed') ?? []
  const closed = myProcedures.data?.filter((procedure) => procedure.state === 'closed') ?? []
  return (
    <>
      <section aria-labelledby="my-procedures">
        <Stack gap="sm">
          <Title order={2} id="my-procedures" size="h4">
            Mis trámites en curso
          </Title>
          {myProcedures.isPending && <Loader aria-label="Cargando los trámites" />}
          {myProcedures.isError && (
            <Alert color="red" role="alert">
              No se pudieron cargar los trámites. Volvé a intentarlo más tarde.
            </Alert>
          )}
          {myProcedures.isSuccess && inProgress.length === 0 && (
            <Text c="dimmed">No hay trámites en curso. Los trámites se inician desde el catálogo.</Text>
          )}
          {inProgress.length > 0 && <MyProcedureGrid procedures={inProgress} />}
        </Stack>
      </section>
      {closed.length > 0 && (
        <section aria-labelledby="my-closed-procedures">
          <Stack gap="sm">
            <Title order={2} id="my-closed-procedures" size="h4">
              Mis trámites cerrados
            </Title>
            <MyProcedureGrid procedures={closed} />
          </Stack>
        </section>
      )}
    </>
  )
}

type Shortcut = { to: string; title: string; text: string }

function ShortcutCard({ shortcut }: { shortcut: Shortcut }) {
  return (
    <Paper component={Link} to={shortcut.to} variant="interactive" withBorder p="lg" td="none" c="inherit">
      <Stack gap="xs">
        <Text fw={600} c="fiuba.7">
          {shortcut.title}
        </Text>
        <Text size="sm" c="dimmed">
          {shortcut.text}
        </Text>
      </Stack>
    </Paper>
  )
}

export function HomePage() {
  const { staff } = useOutletContext<User>()

  const shortcuts: Shortcut[] = [
    ...(staff
      ? [
          {
            to: '/inbox',
            title: 'Bandeja de trámites',
            text: 'Trámites en curso y acciones pendientes de la cuenta.',
          },
        ]
      : []),
    {
      to: '/catalog',
      title: 'Catálogo de trámites',
      text: 'Requisitos, plazos y documentación de cada trámite.',
    },
  ]
  const howTo = staff ? staffHowTo : applicantHowTo

  return (
    <Stack gap="xl">
      <Stack gap="xs">
        <Text size="sm" fw={600} c="fiuba.7">
          Secretaría de Gestión Académica · Facultad de Ingeniería
        </Text>
        <Title order={1}>Sistema Integral de Gestión de Trámites Académicos</Title>
        <Text c="dimmed">
          SIGTA es la plataforma de la Secretaría de Gestión Académica para iniciar, gestionar y seguir los trámites
          académicos de la facultad, del inicio a la resolución.
        </Text>
      </Stack>

      <MyProceduresSection />

      <section aria-labelledby="shortcuts">
        <Stack gap="sm">
          <Title order={2} id="shortcuts" size="h4">
            Accesos
          </Title>
          <SimpleGrid cols={{ base: 1, sm: 2, md: 3 }} spacing="sm">
            {shortcuts.map((shortcut) => (
              <ShortcutCard key={shortcut.to} shortcut={shortcut} />
            ))}
          </SimpleGrid>
        </Stack>
      </section>

      <section aria-labelledby="how">
        <Stack gap="sm">
          <Title order={2} id="how" size="h4">
            {staff ? 'Cómo se trabaja en SIGTA' : 'Cómo se realiza un trámite'}
          </Title>
          <SimpleGrid cols={{ base: 1, md: 3 }} spacing="sm">
            {howTo.map((item, index) => (
              <Paper key={item.title} withBorder p="lg">
                <Stack gap="sm">
                  <ThemeIcon variant="light" radius="xl" aria-hidden>
                    <Text size="sm" fw={700}>
                      {index + 1}
                    </Text>
                  </ThemeIcon>
                  <Title order={3} size="h5">
                    {item.title}
                  </Title>
                  <Text size="sm" c="dimmed">
                    {item.text}
                  </Text>
                </Stack>
              </Paper>
            ))}
          </SimpleGrid>
        </Stack>
      </section>

      <section aria-labelledby="principles">
        <Stack gap="sm">
          <Title order={2} id="principles" size="h4">
            Criterios de la plataforma
          </Title>
          <SimpleGrid cols={{ base: 1, md: 3 }} spacing="lg">
            {principles.map((principle) => (
              <Stack key={principle.title} gap="xs">
                <Title order={3} size="h5">
                  {principle.title}
                </Title>
                <Text size="sm" c="dimmed">
                  {principle.text}
                </Text>
              </Stack>
            ))}
          </SimpleGrid>
        </Stack>
      </section>
    </Stack>
  )
}
