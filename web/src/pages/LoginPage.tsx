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
  Image,
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
import { PageSection } from '../components/PageSection/PageSection'
import { Reveal } from '../components/Reveal/Reveal'
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

// AvailableProcedures lists the procedures already in SIGTA by name, read
// from the catalog, right after the opening: it is what a visitor came to
// find. Their requirements are in the catalog, which the opening leads to.
function AvailableProcedures() {
  const catalog = useQuery(catalogQuery)
  const entries = [...(catalog.data ?? [])].sort((a, b) => a.name.localeCompare(b.name, 'es'))

  return (
    <PageSection id="catalog" title="Trámites disponibles">
      {catalog.isPending && <Skeleton h="var(--sigta-figures-height)" aria-label="Cargando los trámites" />}
      {catalog.isError && <Text c="gray.7">No se pudieron cargar los trámites. Volvé a intentarlo más tarde.</Text>}
      {catalog.isSuccess && entries.length === 0 && <Text c="gray.7">Todavía no hay trámites publicados.</Text>}
      {catalog.isSuccess && entries.length > 0 && (
        <Stack component="ul" gap="md" p={0} m={0}>
          {entries.map((entry) => (
            <Stack key={entry.id} component="li" variant="part" gap="xxs">
              <Text fw={600}>{entry.name}</Text>
              {entry.description !== '' && (
                <Text size="sm" c="gray.7" maw="var(--sigta-measure)" lineClamp={2}>
                  {entry.description}
                </Text>
              )}
            </Stack>
          ))}
        </Stack>
      )}
    </PageSection>
  )
}

// LoginPage is the public landing page. It answers the two questions a
// visitor arrives with: what is needed to start a procedure, and what
// happened to the one already presented. It is white and without boxes, on
// the width of the content from the top bar to the footer: the headline and
// the login on the left, a photograph of the faculty on the right, and below,
// the procedures available by name. What SIGTA is and the criteria it follows
// are in "Acerca de SIGTA".
export function LoginPage() {
  const [params] = useSearchParams()
  const message = loginErrorText(params.get('error'))

  return (
    <AppShell header={{ height: tokens.layout.topBar }}>
      <AppShell.Header>
        <AppHeader
          signature
          contained
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
        <Container py={{ base: 'xl', md: 'xxl' }}>
          <Stack gap="xxxl">
            <Grid component="section" aria-labelledby="hero" gap={{ base: 'xl', lg: 'xxl' }} align="center">
              <Grid.Col span={{ base: 12, md: 7 }}>
                <Stack gap="lg">
                  <Reveal>
                    <Title order={1} id="hero" variant="display" c="fiuba.9">
                      Tus trámites académicos
                    </Title>
                  </Reveal>
                  <Reveal index={1}>
                    <Text size="lg" c="gray.7" maw="var(--sigta-measure)">
                      Consultá qué necesitás para iniciar una solicitud. Si ya la presentaste, podés ver qué área la
                      está revisando y si tenés que completar algo.
                    </Text>
                  </Reveal>
                  {message && <Alert color="red">{message}</Alert>}
                  <Reveal index={2}>
                    <Stack gap="xs">
                      <Group gap="lg">
                        <Button component="a" href="/api/auth/login" size="lg">
                          Ingresar a SIGTA
                        </Button>
                        <Anchor component={Link} to="/catalog" fw={600}>
                          Consultar requisitos
                        </Anchor>
                      </Group>
                      <Text size="sm" c="gray.7">
                        Con tu cuenta institucional de la UBA.
                      </Text>
                    </Stack>
                  </Reveal>
                </Stack>
              </Grid.Col>
              <Grid.Col span={{ base: 12, md: 5 }}>
                <Reveal index={3}>
                  <Image
                    src="/photos/portada-pasillo.jpg"
                    alt="Dos estudiantes repasan un apunte en un pasillo de la facultad."
                  />
                </Reveal>
              </Grid.Col>
            </Grid>

            <Reveal inView>
              <AvailableProcedures />
            </Reveal>
          </Stack>
        </Container>

        <Box component="footer" mt="xxl">
          <Divider />
          <Container py="lg">
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
