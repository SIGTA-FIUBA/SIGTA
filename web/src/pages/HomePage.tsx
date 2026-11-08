import { Paper, SimpleGrid, Stack, Text, ThemeIcon, Title } from '@mantine/core'
import { Link, useOutletContext } from 'react-router'
import type { components } from '../api/schema'
import { applicantSteps, principles, staffSteps } from '../content/platform'

type User = components['schemas']['User']

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
  // The layout passes the person with the session; outside it there is none.
  const user = useOutletContext<User | null | undefined>()
  const staff = user != null && user.role !== 'applicant'

  const shortcuts: Shortcut[] = [
    ...(staff
      ? [
          {
            to: '/inbox',
            title: 'Bandeja de trámites',
            text: 'Trámites en curso y acciones pendientes del perfil.',
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
