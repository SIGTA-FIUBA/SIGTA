import {
  Alert,
  Anchor,
  AppShell,
  Box,
  Button,
  Container,
  Divider,
  Grid,
  Group,
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
import { BrandSignature } from '../components/BrandSignature/BrandSignature'
import { CountUp } from '../components/CountUp/CountUp'
import { GoogleSignInButton } from '../components/GoogleSignInButton/GoogleSignInButton'
import { PageSection } from '../components/PageSection/PageSection'
import { ProcedureFlow } from '../components/ProcedureFlow/ProcedureFlow'
import { Reveal } from '../components/Reveal/Reveal'
import { StepList } from '../components/StepList/StepList'
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

// CatalogSummary says how many procedures are already done in SIGTA, read
// from the catalog, and leads to it. It does not list them: the platform will
// hold more, and the catalog already lists them. The figure counts up when it
// is scrolled to.
function CatalogSummary() {
  const catalog = useQuery(catalogQuery)
  const entries = catalog.data ?? []

  return (
    <PageSection
      id="catalog"
      title="Trámites disponibles"
      lead="Cada uno con sus requisitos, sus plazos y los documentos que pide."
      action={
        <Anchor component={Link} to="/catalog" fw={600}>
          Ver el catálogo completo
        </Anchor>
      }
    >
      {catalog.isPending && <Skeleton h="var(--sigta-figures-height)" aria-label="Cargando los trámites" />}
      {catalog.isError && <Text c="gray.7">No se pudieron cargar los trámites. Volvé a intentarlo más tarde.</Text>}
      {catalog.isSuccess && entries.length === 0 && <Text c="gray.7">Todavía no hay trámites publicados.</Text>}
      {catalog.isSuccess && entries.length > 0 && (
        <Stack gap="xxs">
          <Text component="div" variant="figure" c="fiuba.9">
            <CountUp value={entries.length} />
          </Text>
          <Text c="gray.7">
            {entries.length === 1
              ? 'trámite disponible, con sus requisitos publicados'
              : 'trámites disponibles, con sus requisitos publicados'}
          </Text>
        </Stack>
      )}
    </PageSection>
  )
}

// Principles are the criteria of the platform, in three columns, each under a
// hairline, with no box; there are three, so the row closes.
function Principles() {
  return (
    <PageSection id="principles" title="Cómo trabaja la plataforma">
      <SimpleGrid component="ul" cols={{ base: 1, sm: 3 }} spacing="xl" verticalSpacing="lg" p={0} m={0}>
        {principles.map((principle) => (
          <Stack key={principle.title} component="li" variant="part" gap="xs">
            <Title order={3} size="h4">
              {principle.title}
            </Title>
            <Text size="sm" c="gray.7">
              {principle.text}
            </Text>
          </Stack>
        ))}
      </SimpleGrid>
    </PageSection>
  )
}

// LoginPage is the public landing page: it says what SIGTA is, how a
// procedure is done and how many procedures it holds, and leads to the login.
// It is white and without boxes, on the grid of the home page: the headline
// and the login on the left, the route of a procedure that draws itself on
// the right, and below, the sections with their title in the first quarter,
// which arrive as they are scrolled to.
export function LoginPage() {
  const [params] = useSearchParams()
  const message = loginErrorText(params.get('error'))

  return (
    <AppShell header={{ height: tokens.layout.topBar }}>
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
              <Button component="a" href="/api/auth/login" variant="outline">
                Ingresar
              </Button>
            </>
          }
        />
      </AppShell.Header>

      <AppShell.Main>
        <Container size="lg" py={{ base: 'xl', md: 'xxl' }}>
          <Stack gap="xxxl">
            <Grid component="section" aria-labelledby="hero" gap={{ base: 'xl', lg: 'xxl' }} align="center">
              <Grid.Col span={{ base: 12, md: 7 }}>
                <Stack gap="lg">
                  <Reveal>
                    <Title order={1} id="hero" variant="display" c="fiuba.9">
                      Los trámites académicos, en línea y con seguimiento
                    </Title>
                  </Reveal>
                  <Reveal index={1}>
                    <Text size="lg" c="gray.7" maw="var(--sigta-measure)">
                      Iniciá tu trámite con los requisitos a la vista, seguí en qué instancia está y respondé lo que te
                      pidan. Lo resuelve la Secretaría de Gestión Académica de la FIUBA.
                    </Text>
                  </Reveal>
                  {message && <Alert color="red">{message}</Alert>}
                  <Reveal index={2}>
                    <Stack gap="xs">
                      <Group gap="sm">
                        <GoogleSignInButton />
                        <Button component={Link} to="/catalog" variant="default">
                          Ver el catálogo
                        </Button>
                      </Group>
                      <Text size="sm" c="gray.7">
                        Con tu cuenta institucional de la UBA.
                      </Text>
                    </Stack>
                  </Reveal>
                </Stack>
              </Grid.Col>
              <Grid.Col span={{ base: 12, md: 5 }} visibleFrom="md">
                <ProcedureFlow />
              </Grid.Col>
            </Grid>

            <Reveal inView>
              <PageSection id="how" title="Un solo lugar, del inicio a la resolución">
                <StepList steps={applicantHowTo} />
              </PageSection>
            </Reveal>
            <Reveal inView>
              <CatalogSummary />
            </Reveal>
            <Reveal inView>
              <Principles />
            </Reveal>
          </Stack>
        </Container>

        <Box component="footer" mt="xxl">
          <Divider />
          <Container size="lg" py="lg">
            <Group justify="space-between" gap="md">
              <BrandSignature />
              <Text size="sm" c="gray.7">
                Secretaría de Gestión Académica, Facultad de Ingeniería, Universidad de Buenos Aires
              </Text>
            </Group>
          </Container>
        </Box>
      </AppShell.Main>
    </AppShell>
  )
}
