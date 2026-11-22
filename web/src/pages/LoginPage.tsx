import {
  Alert,
  AppShell,
  Box,
  Button,
  Container,
  Divider,
  Group,
  List,
  Mark,
  SimpleGrid,
  Skeleton,
  Stack,
  Text,
  Title,
} from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import { Link, useSearchParams } from 'react-router'
import { catalogQuery } from '../api/catalog'
import type { components } from '../api/schema'
import { BlueprintPanel } from '../components/BlueprintPanel/BlueprintPanel'
import { BrandSignature } from '../components/BrandSignature/BrandSignature'
import { CountUp } from '../components/CountUp/CountUp'
import { GoogleSignInButton } from '../components/GoogleSignInButton/GoogleSignInButton'
import { ProcedureFlow } from '../components/ProcedureFlow/ProcedureFlow'
import { Reveal } from '../components/Reveal/Reveal'
import { applicantHowTo, principles } from '../content/platform'
import { AppHeader } from '../layout/AppHeader'
import { tokens } from '../theme/tokens'

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

// HowItWorks is how a procedure is done, in three steps on a line that draws
// itself in the action color as they are scrolled to; each step arrives after
// the one before.
function HowItWorks() {
  return (
    <Box component="section" bg="gray.0" aria-labelledby="how">
      <Container size="lg" py="xxxl">
        <Stack gap="xxl">
          <Reveal inView>
            <Stack gap="xs">
              <Text variant="overline" c="fiuba.7">
                Cómo funciona
              </Text>
              <Title order={2} id="how" variant="section">
                Un solo lugar, del inicio a la resolución
              </Title>
            </Stack>
          </Reveal>
          <List type="ordered" variant="steps">
            {applicantHowTo.map((step, index) => (
              <List.Item key={step.title}>
                <Reveal index={index} inView>
                  <Stack gap="sm">
                    <Text variant="figure" component="span" c="fiuba.7" aria-hidden>
                      {String(index + 1).padStart(2, '0')}
                    </Text>
                    <Title order={3}>{step.title}</Title>
                    <Text size="sm" c="dimmed">
                      {step.text}
                    </Text>
                  </Stack>
                </Reveal>
              </List.Item>
            ))}
          </List>
        </Stack>
      </Container>
    </Box>
  )
}

// CatalogSummary says how many procedures are already done in SIGTA, read
// from the catalog, and leads to it. It does not list them: the platform will
// hold more, and the catalog already lists them. The figure counts up when it
// is scrolled to.
function CatalogSummary() {
  const catalog = useQuery(catalogQuery)
  const entries = catalog.data ?? []

  return (
    <Box component="section" aria-labelledby="catalog">
      <Container size="lg" py="xxxl">
        <Reveal inView>
          <BlueprintPanel radius="lg" p={{ base: 'lg', md: 'xxl' }}>
            <SimpleGrid cols={{ base: 1, md: 2 }} spacing="xl">
              <Stack gap="lg" align="flex-start">
                <Stack gap="xs">
                  <Text variant="overline" c="fiuba.1">
                    Catálogo
                  </Text>
                  <Title order={2} id="catalog" c="white" variant="section">
                    Trámites disponibles
                  </Title>
                </Stack>
                <Button component={Link} to="/catalog" variant="white" radius="xl">
                  Ver el catálogo completo
                </Button>
              </Stack>
              <Stack justify="center">
                {catalog.isPending && <Skeleton h="var(--sigta-figures-height)" aria-label="Cargando los trámites" />}
                {catalog.isError && (
                  <Text c="fiuba.1">No se pudieron cargar los trámites. Volvé a intentarlo más tarde.</Text>
                )}
                {catalog.isSuccess && entries.length === 0 && (
                  <Text c="fiuba.1">Todavía no hay trámites publicados.</Text>
                )}
                {catalog.isSuccess && entries.length > 0 && (
                  <Stack gap="xxs">
                    <Text component="div" variant="figure" c="white">
                      <CountUp value={entries.length} />
                    </Text>
                    <Text size="sm" c="fiuba.1">
                      {entries.length === 1
                        ? 'trámite disponible, con sus requisitos publicados'
                        : 'trámites disponibles, con sus requisitos publicados'}
                    </Text>
                  </Stack>
                )}
              </Stack>
            </SimpleGrid>
          </BlueprintPanel>
        </Reveal>
      </Container>
    </Box>
  )
}

function Principles() {
  return (
    <Box component="section" bg="gray.0" aria-labelledby="principles">
      <Container size="lg" py="xxxl">
        <Stack gap="xxl">
          <Reveal inView>
            <Stack gap="xs">
              <Text variant="overline" c="fiuba.7">
                Criterios
              </Text>
              <Title order={2} id="principles" variant="section">
                Cómo trabaja la plataforma
              </Title>
            </Stack>
          </Reveal>
          <SimpleGrid cols={{ base: 1, md: 3 }} spacing="xl">
            {principles.map((principle, index) => (
              <Reveal key={principle.title} index={index} inView>
                <Stack gap="xs">
                  <Divider size="sm" color="fiuba.7" />
                  <Title order={3}>{principle.title}</Title>
                  <Text size="sm" c="dimmed">
                    {principle.text}
                  </Text>
                </Stack>
              </Reveal>
            ))}
          </SimpleGrid>
        </Stack>
      </Container>
    </Box>
  )
}

// LoginPage is the public landing page: it says what SIGTA is, how a
// procedure is done and how many procedures it holds, and leads to the login.
// The first screen is the blueprint, with the route of a procedure that draws
// itself; the rest arrives as it is scrolled to.
export function LoginPage() {
  const [params] = useSearchParams()
  const message = loginErrorText(params.get('error'))

  return (
    <AppShell header={{ height: tokens.layout.topBar }} variant="landing">
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
        <BlueprintPanel component="section" radius={0} aria-labelledby="hero">
          <Container size="lg" py={{ base: 'xxl', md: 'xxxl' }}>
            <SimpleGrid cols={{ base: 1, md: 2 }} spacing="xxl">
              <Stack gap="lg" justify="center">
                <Reveal>
                  <Text variant="overline" c="fiuba.1">
                    Secretaría de Gestión Académica, FIUBA
                  </Text>
                </Reveal>
                <Reveal index={1}>
                  <Title order={1} id="hero" variant="display" c="white">
                    Los trámites académicos,{' '}
                    <Mark color="fiuba.6" data-sweep>
                      en línea
                    </Mark>{' '}
                    y con seguimiento
                  </Title>
                </Reveal>
                <Reveal index={2}>
                  <Text size="lg" c="fiuba.1" maw="var(--sigta-measure)">
                    Iniciá tu trámite con los requisitos a la vista, seguí en qué instancia está y respondé lo que te
                    pidan.
                  </Text>
                </Reveal>
                {message && <Alert color="red">{message}</Alert>}
                <Reveal index={3}>
                  <Stack gap="xs">
                    <Group gap="sm">
                      <GoogleSignInButton />
                      <Button component={Link} to="/catalog" variant="outline" color="white" radius="xl">
                        Ver el catálogo
                      </Button>
                    </Group>
                    <Text size="sm" c="fiuba.1">
                      Con tu cuenta institucional de la UBA.
                    </Text>
                  </Stack>
                </Reveal>
              </Stack>
              <Box visibleFrom="md">
                <ProcedureFlow />
              </Box>
            </SimpleGrid>
          </Container>
        </BlueprintPanel>

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
