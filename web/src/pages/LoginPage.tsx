import {
  Alert,
  AppShell,
  Badge,
  Box,
  Button,
  Container,
  Group,
  Paper,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
  Timeline,
  Title,
} from '@mantine/core'
import { Link, useSearchParams } from 'react-router'
import type { components } from '../api/schema'
import { GoogleSignInButton } from '../components/GoogleSignInButton/GoogleSignInButton'
import { applicantSteps, pillars, principles } from '../content/platform'
import { AppHeader } from '../layout/AppHeader'

type LoginError = components['schemas']['LoginError']

// The callback only says why a login failed; the text lives here, so the
// language of the interface does not depend on the backend.
const messages: Record<LoginError, string> = {
  state_mismatch: 'El ingreso no coincide con el que se inició. Volvé a intentarlo.',
  rejected: 'El ingreso no se pudo completar. Volvé a intentarlo.',
  domain_not_allowed: 'Ingresá con tu cuenta institucional.',
}

function messageFor(error: string | null): string | undefined {
  if (error === null) {
    return undefined
  }
  return Object.hasOwn(messages, error) ? messages[error as LoginError] : messages.rejected
}

// TrackingPreview shows what following a procedure looks like, with the
// platform's own components and an invented procedure.
function TrackingPreview() {
  return (
    <Paper shadow="md" p="lg" radius="md" c="black" aria-label="Ejemplo del seguimiento de un trámite">
      <Stack gap="md">
        <Group justify="space-between" align="flex-start" gap="xs">
          <Stack gap={0}>
            <Text fw={600}>Equivalencia entre asignaturas</Text>
            <Text size="sm" c="dimmed">
              Iniciado hace 24 días
            </Text>
          </Stack>
          <Badge variant="light" color="orange">
            En evaluación externa
          </Badge>
        </Group>
        <Timeline active={2} bulletSize={20} lineWidth={2}>
          <Timeline.Item title="Solicitud presentada">
            <Text size="xs" c="dimmed">
              Hace 24 días
            </Text>
          </Timeline.Item>
          <Timeline.Item title="Apertura del expediente">
            <Text size="xs" c="dimmed">
              Dirección de Alumnos · hace 20 días
            </Text>
          </Timeline.Item>
          <Timeline.Item title="Dictamen de la Comisión Curricular">
            <Text size="xs" c="dimmed">
              Instancia actual · 15 días en evaluación
            </Text>
          </Timeline.Item>
          <Timeline.Item title="Resolución y registro en el SIU Guaraní" />
        </Timeline>
      </Stack>
    </Paper>
  )
}

// LoginPage is the public landing page: it says what SIGTA is, which
// procedures it covers and how they are done, and leads to the login.
export function LoginPage() {
  const [params] = useSearchParams()
  const message = messageFor(params.get('error'))

  return (
    <AppShell header={{ height: 64 }}>
      <AppShell.Header>
        <AppHeader
          end={
            <Button component={Link} to="/catalog" variant="subtle">
              Catálogo de trámites
            </Button>
          }
        />
      </AppShell.Header>

      <AppShell.Main>
        <Box component="section" bg="fiuba.8" c="white" aria-labelledby="hero">
          <Container size="lg" py="xl" mih="calc(100dvh - var(--app-shell-header-height))" display="flex">
            <SimpleGrid cols={{ base: 1, md: 2 }} spacing="xl" py="xl" my="auto" w="100%">
              <Stack gap="lg" justify="center">
                <Text size="sm" fw={600} tt="uppercase">
                  Secretaría de Gestión Académica · FIUBA
                </Text>
                <Title order={1} id="hero" variant="display">
                  Los trámites académicos de la facultad, en línea y con seguimiento
                </Title>
                <Text size="lg">
                  Los trámites de la Secretaría de Gestión Académica se inician desde SIGTA con la cuenta institucional
                  de la UBA, y cada instancia queda a la vista hasta la resolución.
                </Text>
                {message && (
                  <Alert color="red" variant="white">
                    {message}
                  </Alert>
                )}
                <Group gap="md">
                  <GoogleSignInButton />
                  <Button component={Link} to="/catalog" variant="white" size="md">
                    Consultar el catálogo
                  </Button>
                </Group>
              </Stack>
              <Stack justify="center">
                <TrackingPreview />
              </Stack>
            </SimpleGrid>
          </Container>
        </Box>

        <Box component="section" aria-labelledby="core">
          <Container size="lg" py="xl">
            <Stack gap="lg" py="xl">
              <Stack gap="xs">
                <Title order={2} id="core">
                  Qué resuelve SIGTA
                </Title>
                <Text c="dimmed">
                  Una sola plataforma para los trámites académicos de la Secretaría, del inicio a la resolución.
                </Text>
              </Stack>
              <SimpleGrid cols={{ base: 1, md: 3 }} spacing="md">
                {pillars.map((pillar) => (
                  <Paper key={pillar.title} withBorder p="lg">
                    <Stack gap="xs">
                      <Title order={3} size="h4">
                        {pillar.title}
                      </Title>
                      <Text c="dimmed">{pillar.text}</Text>
                    </Stack>
                  </Paper>
                ))}
              </SimpleGrid>
            </Stack>
          </Container>
        </Box>

        <Box component="section" bg="gray.0" aria-labelledby="how">
          <Container size="lg" py="xl">
            <Stack gap="lg" py="xl">
              <Title order={2} id="how">
                Cómo se realiza un trámite
              </Title>
              <SimpleGrid cols={{ base: 1, md: 3 }} spacing="md">
                {applicantSteps.map((step, index) => (
                  <Stack key={step.title} gap="sm">
                    <ThemeIcon variant="filled" radius="xl" size="lg" aria-hidden>
                      <Text fw={700}>{index + 1}</Text>
                    </ThemeIcon>
                    <Title order={3} size="h4">
                      {step.title}
                    </Title>
                    <Text c="dimmed">{step.text}</Text>
                  </Stack>
                ))}
              </SimpleGrid>
            </Stack>
          </Container>
        </Box>

        <Box component="section" aria-labelledby="principles">
          <Container size="lg" py="xl">
            <Stack gap="lg" py="xl">
              <Title order={2} id="principles">
                Criterios de la plataforma
              </Title>
              <SimpleGrid cols={{ base: 1, md: 3 }} spacing="xl">
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
          </Container>
        </Box>

        <Box component="footer" bg="fiuba.9" c="white">
          <Container size="lg" py="lg">
            <Text size="sm">
              Secretaría de Gestión Académica · Facultad de Ingeniería · Universidad de Buenos Aires
            </Text>
          </Container>
        </Box>
      </AppShell.Main>
    </AppShell>
  )
}
