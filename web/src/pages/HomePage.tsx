import { Alert, Badge, Group, Loader, Paper, SimpleGrid, Stack, Text, ThemeIcon, Title } from '@mantine/core'
import { Link, useOutletContext } from 'react-router'
import { useMyProcedures } from '../api/procedures'
import type { components } from '../api/schema'
import { applicantSteps, principles, staffSteps } from '../content/platform'
import { stateColors, stateNames } from '../procedureState'

type User = components['schemas']['User']
type MyProcedure = components['schemas']['MyProcedure']

const date = new Intl.DateTimeFormat('es-AR', {
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
        <Text size="sm">Paso actual: {procedure.step_name}</Text>
        <Text size="sm" c="dimmed">
          Último movimiento: {date.format(new Date(procedure.last_moved_at))}
        </Text>
      </Stack>
    </Paper>
  )
}

// MyProcedures lists what the person started and is not closed yet, so that
// signing in answers where each procedure is.
function MyProcedures() {
  const mine = useMyProcedures()
  return (
    <section aria-labelledby="mine">
      <Stack gap="sm">
        <Title order={2} id="mine" size="h4">
          Mis trámites en curso
        </Title>
        {mine.isPending && <Loader aria-label="Cargando los trámites en curso" />}
        {mine.isError && (
          <Alert color="red" role="alert">
            No se pudieron cargar los trámites en curso. Volvé a intentarlo más tarde.
          </Alert>
        )}
        {mine.isSuccess && mine.data.length === 0 && (
          <Text c="dimmed">No hay trámites en curso. Los trámites se inician desde el catálogo.</Text>
        )}
        {mine.isSuccess && mine.data.length > 0 && (
          <SimpleGrid cols={{ base: 1, sm: 2, md: 3 }} spacing="sm">
            {mine.data.map((procedure) => (
              <MyProcedureCard key={procedure.id} procedure={procedure} />
            ))}
          </SimpleGrid>
        )}
      </Stack>
    </section>
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
  const steps = staff ? staffSteps : applicantSteps

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

      <MyProcedures />

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
            {steps.map((step, index) => (
              <Paper key={step.title} withBorder p="lg">
                <Stack gap="sm">
                  <ThemeIcon variant="light" radius="xl" aria-hidden>
                    <Text size="sm" fw={700}>
                      {index + 1}
                    </Text>
                  </ThemeIcon>
                  <Title order={3} size="h5">
                    {step.title}
                  </Title>
                  <Text size="sm" c="dimmed">
                    {step.text}
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
