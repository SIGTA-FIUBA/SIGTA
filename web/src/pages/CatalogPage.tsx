import { Alert, Button, Grid, Skeleton, Stack, Text, Title } from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router'
import { catalogQuery } from '../api/catalog'
import type { components } from '../api/schema'
import { Fact } from '../components/Fact/Fact'
import { Facts } from '../components/Facts/Facts'
import { documentDescription } from '../content/documentFormats'

type CatalogEntry = components['schemas']['CatalogEntry']
type DocumentRequirement = components['schemas']['DocumentRequirement']

function requiredText(required: boolean) {
  return required ? 'Obligatorio' : 'Opcional'
}

// A version of an older format has no accepted types or size, so each part is
// shown only when it is known.
function documentText(document: DocumentRequirement) {
  return [`${document.name} (${requiredText(document.required)}).`, documentDescription(document)]
    .filter(Boolean)
    .join(' ')
}

// Lines is what goes in a fact of the catalog: one line each, with no bullet.
function Lines({ items }: { items: string[] }) {
  return (
    <Stack gap="xxs">
      {items.map((item, index) => (
        <Text key={index} size="sm">
          {item}
        </Text>
      ))}
    </Stack>
  )
}

// CatalogItem is one procedure type, as an entry of an index: a rule above,
// the name, what it is and how to start it in the first third, and what it
// asks for, as facts between hairlines, in the rest. Each fact is shown only
// when the type has something to say in it.
function CatalogItem({ entry }: { entry: CatalogEntry }) {
  const facts = [
    { label: 'Datos', items: entry.fields.map((field) => `${field.name} (${requiredText(field.required)})`) },
    { label: 'Documentos', items: entry.documents.map(documentText) },
    { label: 'Plazos', items: entry.deadlines },
    { label: 'Condiciones de admisibilidad', items: entry.admissibility },
  ].filter((fact) => fact.items.length > 0)
  return (
    <Stack component="article" variant="part" gap={0} aria-label={entry.name}>
      <Grid gap={{ base: 'lg', lg: 'xl' }}>
        <Grid.Col span={{ base: 12, lg: 4 }}>
          <Stack gap="md" align="flex-start">
            <Title order={2}>{entry.name}</Title>
            {entry.description !== '' && (
              <Text size="sm" c="gray.7" maw="var(--sigta-measure)">
                {entry.description}
              </Text>
            )}
            <Button component={Link} to={`/catalog/${entry.id}/start`} aria-label={`Iniciar ${entry.name}`}>
              Iniciar trámite
            </Button>
          </Stack>
        </Grid.Col>
        <Grid.Col span={{ base: 12, lg: 8 }}>
          {facts.length === 0 ? (
            <Text size="sm" c="gray.7">
              No pide datos ni documentos.
            </Text>
          ) : (
            <Facts label={`Qué pide ${entry.name}`}>
              {facts.map((fact) => (
                <Fact key={fact.label} label={fact.label} split>
                  <Lines items={fact.items} />
                </Fact>
              ))}
            </Facts>
          )}
        </Grid.Col>
      </Grid>
    </Stack>
  )
}

export function CatalogPage() {
  const catalog = useQuery(catalogQuery)

  const entries = [...(catalog.data ?? [])].sort((a, b) => a.name.localeCompare(b.name, 'es'))
  return (
    <Stack gap="xxl">
      <Stack gap="xs">
        <Title order={1} variant="display" c="fiuba.9">
          Catálogo de trámites
        </Title>
        <Text c="gray.7" maw="var(--sigta-measure)">
          Requisitos, plazos y documentación de cada trámite disponible.
        </Text>
      </Stack>
      {catalog.isPending && (
        <Stack gap="sm" aria-label="Cargando los trámites">
          <Skeleton h="var(--sigta-row-height)" />
          <Skeleton h="var(--sigta-row-height)" />
        </Stack>
      )}
      {catalog.isError && (
        <Alert color="red" role="alert">
          No pudimos cargar los trámites. Volvé a intentarlo más tarde.
        </Alert>
      )}
      {catalog.isSuccess && entries.length === 0 && <Text c="dimmed">Todavía no hay trámites disponibles.</Text>}
      {entries.length > 0 && (
        <Stack gap="xxl">
          {entries.map((entry) => (
            <CatalogItem key={entry.id} entry={entry} />
          ))}
        </Stack>
      )}
    </Stack>
  )
}
