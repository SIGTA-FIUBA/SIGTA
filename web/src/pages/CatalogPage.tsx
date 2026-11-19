import { Alert, Button, Group, List, Loader, Stack, Text, Title } from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router'
import { catalogQuery } from '../api/catalog'
import type { components } from '../api/schema'
import { documentDescription } from '../documentFormats'

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

function Section({ title, items }: { title: string; items: string[] }) {
  if (items.length === 0) {
    return null
  }
  return (
    <Stack gap="xs">
      <Title order={3}>{title}</Title>
      <List>
        {items.map((item, index) => (
          <List.Item key={index}>{item}</List.Item>
        ))}
      </List>
    </Stack>
  )
}

function Entry({ entry }: { entry: CatalogEntry }) {
  return (
    <Stack component="article" gap="sm">
      <Title order={2}>{entry.name}</Title>
      {entry.description !== '' && <Text>{entry.description}</Text>}
      <Section title="Datos" items={entry.fields.map((field) => `${field.name} (${requiredText(field.required)})`)} />
      <Section title="Documentos" items={entry.documents.map(documentText)} />
      <Section title="Plazos" items={entry.deadlines} />
      <Section title="Condiciones de admisibilidad" items={entry.admissibility} />
      <Group>
        <Button component={Link} to={`/catalog/${entry.id}/start`} variant="light" aria-label={`Iniciar ${entry.name}`}>
          Iniciar trámite
        </Button>
      </Group>
    </Stack>
  )
}

export function CatalogPage() {
  const catalog = useQuery(catalogQuery)

  const entries = [...(catalog.data ?? [])].sort((a, b) => a.name.localeCompare(b.name, 'es'))
  return (
    <Stack gap="lg">
      <Stack gap={0}>
        <Title order={1}>Catálogo de trámites</Title>
        <Text c="dimmed">Requisitos, plazos y documentación de cada trámite disponible.</Text>
      </Stack>
      {catalog.isPending && <Loader aria-label="Cargando los trámites" />}
      {catalog.isError && (
        <Alert color="red" role="alert">
          No pudimos cargar los trámites. Volvé a intentarlo más tarde.
        </Alert>
      )}
      {catalog.isSuccess && entries.length === 0 && <Text c="dimmed">Todavía no hay trámites disponibles.</Text>}
      {entries.map((entry) => (
        <Entry key={entry.id} entry={entry} />
      ))}
    </Stack>
  )
}
