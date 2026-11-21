import {
  Alert,
  AppShell,
  Badge,
  Box,
  Button,
  Container,
  Divider,
  Grid,
  Group,
  Image,
  Mark,
  Paper,
  SimpleGrid,
  Skeleton,
  Stack,
  Text,
  Timeline,
  Title,
} from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import { Link, useSearchParams } from 'react-router'
import { catalogQuery } from '../api/catalog'
import type { components } from '../api/schema'
import { ActionCard } from '../components/ActionCard/ActionCard'
import { BrandSignature } from '../components/BrandSignature/BrandSignature'
import { GoogleSignInButton } from '../components/GoogleSignInButton/GoogleSignInButton'
import { StatusBadge } from '../components/StatusBadge/StatusBadge'
import { applicantHowTo, principles } from '../content/platform'
import { AppHeader } from '../layout/AppHeader'

type LoginError = components['schemas']['LoginError']

// The callback only says why a login failed; the text lives here, so the
// language of the interface does not depend on the backend.
const loginErrorTexts: Record<LoginError, string> = {
  state_mismatch: 'El ingreso no coincide con el que se inició. Volvé a intentarlo.',
  rejected: 'El ingreso no se pudo completar. Volvé a intentarlo.',
  domain_not_allowed: 'Ingresá con tu cuenta institucional.',
}

function loginErrorText(error: string | null): string | undefined {
  if (error === null) {
    return undefined
  }
  return Object.hasOwn(loginErrorTexts, error) ? loginErrorTexts[error as LoginError] : loginErrorTexts.rejected
}

// The catalog shows its first entries; the rest are a click away.
const catalogPreview = 4

// ProcedureCollage takes the place of a photo: pieces of the platform itself,
// with an invented procedure, on tiles of the tints of the cyan. It only
// illustrates what the text says, so screen readers skip it.
function ProcedureCollage() {
  return (
    <SimpleGrid cols={2} spacing="md" visibleFrom="md" aria-hidden>
      <Paper bg="fiuba.0" radius="lg" p="lg">
        <Stack gap="sm">
          <Text variant="overline" c="fiuba.9">
            Mi trámite
          </Text>
          <Text fw={600} size="lg">
            Equivalencia entre asignaturas
          </Text>
          <Box>
            <StatusBadge tone="office">En la Secretaría</StatusBadge>
          </Box>
          <Timeline active={1} bulletSize={16} lineWidth={2}>
            <Timeline.Item title="Solicitud presentada" />
            <Timeline.Item title="Revisión de la Secretaría" />
            <Timeline.Item title="Dictamen del área" />
          </Timeline>
        </Stack>
      </Paper>
      <Paper bg="fiuba.9" radius="lg" p="xl" display="flex">
        <Image src="/brand/sigta-isotipo-white.svg" alt="" w="60%" h="auto" m="auto" />
      </Paper>
      <Paper bg="fiuba.1" radius="lg" p="lg">
        <Stack gap="sm" align="flex-start">
          <Text variant="overline" c="fiuba.9">
            Acción requerida
          </Text>
          <Text fw={600}>Falta el programa de la asignatura</Text>
          <Badge variant="filled" color="fiuba.9">
            Responder
          </Badge>
        </Stack>
      </Paper>
      <Paper withBorder radius="lg" p="lg">
        <Stack gap="sm" align="flex-start">
          <Text variant="overline" c="dimmed">
            Cierre
          </Text>
          <StatusBadge tone="closed">Cerrado</StatusBadge>
          <Text size="sm" c="dimmed">
            Dictamen adjunto y comunicado al titular
          </Text>
        </Stack>
      </Paper>
    </SimpleGrid>
  )
}

// The three steps of a procedure go on white, on the tint and on the navy,
// from the first to the last.
const stepTiles = [
  { bg: 'white', c: undefined, dimmed: 'dimmed', withBorder: true },
  { bg: 'fiuba.1', c: undefined, dimmed: undefined, withBorder: false },
  { bg: 'fiuba.9', c: 'white', dimmed: 'fiuba.1', withBorder: false },
]

function HowItWorks() {
  return (
    <Box component="section" bg="gray.0" aria-labelledby="how">
      <Container size="lg" py="xxxl">
        <Stack gap="xl">
          <Stack gap="xs">
            <Text variant="overline" c="fiuba.7">
              Cómo funciona
            </Text>
            <Title order={2} id="how" variant="section">
              Un solo lugar, del inicio a la resolución
            </Title>
          </Stack>
          <SimpleGrid cols={{ base: 1, md: 3 }} spacing="md">
            {applicantHowTo.map((step, index) => {
              const tile = stepTiles[index % stepTiles.length]
              return (
                <Paper key={step.title} bg={tile.bg} c={tile.c} withBorder={tile.withBorder} radius="lg" p="xl">
                  <Stack gap="sm">
                    <Text variant="figure" component="span" aria-hidden>
                      {String(index + 1).padStart(2, '0')}
                    </Text>
                    <Title order={3} c={tile.c}>
                      {step.title}
                    </Title>
                    <Text size="sm" c={tile.dimmed}>
                      {step.text}
                    </Text>
                  </Stack>
                </Paper>
              )
            })}
          </SimpleGrid>
        </Stack>
      </Container>
    </Box>
  )
}

// CatalogSummary says how many procedures are already done in SIGTA and shows
// the first ones, from the catalog itself: nothing on this page is written by
// hand about a procedure.
function CatalogSummary() {
  const catalog = useQuery(catalogQuery)
  const entries = [...(catalog.data ?? [])].sort((a, b) => a.name.localeCompare(b.name, 'es'))

  return (
    <Box component="section" aria-labelledby="catalog">
      <Container size="lg" py="xxxl">
        <Grid gap="md">
          <Grid.Col span={{ base: 12, md: 4 }}>
            <Paper bg="fiuba.9" c="white" radius="lg" p="xl" h="100%">
              <Stack gap="xl" justify="space-between" h="100%">
                <Stack gap="xs">
                  <Text variant="overline" c="fiuba.1">
                    Catálogo
                  </Text>
                  <Title order={2} id="catalog" c="white" variant="section">
                    Los trámites que ya se hacen en SIGTA
                  </Title>
                </Stack>
                {catalog.isPending && <Skeleton h="var(--sigta-figures-height)" aria-label="Cargando los trámites" />}
                {catalog.isError && (
                  <Text c="fiuba.1">No pudimos cargar los trámites. Volvé a intentarlo más tarde.</Text>
                )}
                {catalog.isSuccess && entries.length === 0 && (
                  <Text c="fiuba.1">Todavía no hay trámites publicados.</Text>
                )}
                {catalog.isSuccess && entries.length > 0 && (
                  <Stack gap="xxs">
                    <Text variant="figure" c="fiuba.1">
                      {entries.length}
                    </Text>
                    <Text size="sm">
                      {entries.length === 1
                        ? 'trámite disponible, con sus requisitos publicados'
                        : 'trámites disponibles, con sus requisitos publicados'}
                    </Text>
                  </Stack>
                )}
                <Box>
                  <Button component={Link} to="/catalog" variant="white" radius="xl">
                    Ver el catálogo completo
                  </Button>
                </Box>
              </Stack>
            </Paper>
          </Grid.Col>
          <Grid.Col span={{ base: 12, md: 8 }}>
            <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="md">
              {catalog.isPending &&
                Array.from({ length: catalogPreview }, (_, index) => (
                  <Skeleton key={index} h="var(--sigta-row-height)" />
                ))}
              {entries.slice(0, catalogPreview).map((entry) => (
                <ActionCard key={entry.id} to="/catalog" title={entry.name} description={entry.description} />
              ))}
            </SimpleGrid>
          </Grid.Col>
        </Grid>
      </Container>
    </Box>
  )
}

function Principles() {
  return (
    <Box component="section" bg="gray.0" aria-labelledby="principles">
      <Container size="lg" py="xxxl">
        <Stack gap="xl">
          <Stack gap="xs">
            <Text variant="overline" c="fiuba.7">
              Criterios
            </Text>
            <Title order={2} id="principles" variant="section">
              Cómo trabaja la plataforma
            </Title>
          </Stack>
          <SimpleGrid cols={{ base: 1, md: 3 }} spacing="xl">
            {principles.map((principle) => (
              <Stack key={principle.title} gap="xs">
                <Divider size="sm" color="fiuba.7" />
                <Title order={3}>{principle.title}</Title>
                <Text size="sm" c="dimmed">
                  {principle.text}
                </Text>
              </Stack>
            ))}
          </SimpleGrid>
        </Stack>
      </Container>
    </Box>
  )
}

// LoginPage is the public landing page: it says what SIGTA is, how a
// procedure is done and which procedures it holds, and leads to the login.
export function LoginPage() {
  const [params] = useSearchParams()
  const message = loginErrorText(params.get('error'))

  return (
    <AppShell header={{ height: 64 }} variant="landing">
      <AppShell.Header>
        <AppHeader
          signature
          end={
            <>
              <Button component={Link} to="/catalog" variant="subtle" visibleFrom="sm">
                Catálogo de trámites
              </Button>
              <Button component={Link} to="/about" variant="subtle" visibleFrom="sm">
                Acerca de SIGTA
              </Button>
              <Button component="a" href="/api/auth/login" variant="outline" radius="xl">
                Ingresar
              </Button>
            </>
          }
        />
      </AppShell.Header>

      <AppShell.Main>
        <Box component="section" aria-labelledby="hero">
          <Container size="lg" py="xxxl">
            <SimpleGrid cols={{ base: 1, md: 2 }} spacing="xxl">
              <Stack gap="lg" justify="center">
                <Text variant="overline" c="dimmed">
                  Secretaría de Gestión Académica, FIUBA
                </Text>
                <Title order={1} id="hero" variant="display">
                  Tus trámites académicos, <Mark>en línea</Mark> y con seguimiento
                </Title>
                <Text size="lg" c="dimmed" maw="var(--sigta-measure)">
                  Iniciá tu trámite con los requisitos a la vista, seguí en qué instancia está y respondé lo que te
                  pidan, sin correos ni papeles.
                </Text>
                {message && <Alert color="red">{message}</Alert>}
                <Stack gap="xs">
                  <Group gap="sm">
                    <GoogleSignInButton />
                    <Button component={Link} to="/catalog" variant="outline" radius="xl">
                      Ver el catálogo
                    </Button>
                  </Group>
                  <Text size="sm" c="dimmed">
                    Con tu cuenta institucional de la UBA.
                  </Text>
                </Stack>
              </Stack>
              <ProcedureCollage />
            </SimpleGrid>
          </Container>
        </Box>

        <HowItWorks />
        <CatalogSummary />
        <Principles />

        {/* The footer is white: the FIUBA logo is lost on cyan or navy. */}
        <Box component="footer" bg="white">
          <Divider />
          <Container size="lg" py="lg">
            <Group justify="space-between" gap="md">
              <BrandSignature />
              <Text size="sm" c="dimmed">
                Secretaría de Gestión Académica, Facultad de Ingeniería, Universidad de Buenos Aires
              </Text>
            </Group>
          </Container>
        </Box>
      </AppShell.Main>
    </AppShell>
  )
}
