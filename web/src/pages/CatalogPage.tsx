import { Alert, Container, List, Loader, Stack, Text, Title } from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import type { components } from '../api/schema'

type CatalogEntry = components['schemas']['CatalogEntry']
type DocumentRequirement = components['schemas']['DocumentRequirement']

const typeNames: Record<string, string> = {
  'application/pdf': 'PDF',
  'image/jpeg': 'JPG',
  'image/png': 'PNG',
}

function typeName(mediaType: string) {
  return typeNames[mediaType] ?? mediaType.split('/').pop()!.toUpperCase()
}

function sizeText(bytes: number) {
  const megabytes = bytes / (1024 * 1024)
  if (megabytes >= 1) {
    return `${Number(megabytes.toFixed(1))} MB`
  }
  return `${Math.ceil(bytes / 1024)} KB`
}

function requiredText(required: boolean) {
  return required ? 'Obligatorio' : 'Opcional'
}

// A version of an older format has no accepted types or size, so each part is
// shown only when it is known.
function documentText(document: DocumentRequirement) {
  const parts = [`${document.name} (${requiredText(document.required)}).`]
  if (document.accepted_types.length > 0) {
    parts.push(`Formatos: ${document.accepted_types.map(typeName).join(', ')}.`)
  }
  if (document.max_size > 0) {
    parts.push(`Tamaño máximo: ${sizeText(document.max_size)}.`)
  }
  return parts.join(' ')
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
    </Stack>
  )
}

export function CatalogPage() {
  const catalog = useQuery({
    queryKey: ['procedure-types'],
    queryFn: async () => {
      const { data, error } = await api.GET('/procedure-types')
      if (error !== undefined || data === undefined) {
        throw new Error('The catalog could not be loaded')
      }
      return data
    },
  })

  const entries = [...(catalog.data ?? [])].sort((a, b) => a.name.localeCompare(b.name, 'es'))
  return (
    <Container size="md" py="md">
      <Stack gap="lg">
        <Title order={1}>Trámites</Title>
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
    </Container>
  )
}
