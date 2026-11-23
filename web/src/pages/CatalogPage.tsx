import { Accordion, Alert, Button, Skeleton, Stack, Text, Title } from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router'
import { catalogQuery } from '../api/catalog'
import type { components } from '../api/schema'
import { Fact } from '../components/Fact/Fact'
import { Facts } from '../components/Facts/Facts'
import { PageHeader } from '../components/PageHeader/PageHeader'

type CatalogEntry = components['schemas']['CatalogEntry']

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

// CatalogItem is one procedure type, folded: the name and what it is. Unfolded
// it says who can ask for it and its deadlines, and offers to start it. What
// the form asks for (fields, documents, formats and sizes) is left to the form
// itself: the catalog is for understanding what each procedure is.
function CatalogItem({ entry }: { entry: CatalogEntry }) {
  const facts = [
    { label: 'Quién puede pedirlo', items: entry.admissibility },
    { label: 'Plazos', items: entry.deadlines },
  ].filter((fact) => fact.items.length > 0)
  return (
    <Accordion.Item value={String(entry.id)}>
      <Accordion.Control>
        <Stack gap="xxs">
          <Title order={2} size="h4">
            {entry.name}
          </Title>
          {entry.description !== '' && (
            <Text size="sm" c="gray.7" maw="var(--sigta-measure)">
              {entry.description}
            </Text>
          )}
        </Stack>
      </Accordion.Control>
      <Accordion.Panel>
        <Stack gap="lg" align="flex-start">
          {facts.length > 0 && (
            <Facts label={`Sobre ${entry.name}`}>
              {facts.map((fact) => (
                <Fact key={fact.label} label={fact.label} split>
                  <Lines items={fact.items} />
                </Fact>
              ))}
            </Facts>
          )}
          <Button component={Link} to={`/catalog/${entry.id}/start`} aria-label={`Iniciar ${entry.name}`}>
            Iniciar trámite
          </Button>
        </Stack>
      </Accordion.Panel>
    </Accordion.Item>
  )
}

export function CatalogPage() {
  const catalog = useQuery(catalogQuery)

  const entries = [...(catalog.data ?? [])].sort((a, b) => a.name.localeCompare(b.name, 'es'))
  return (
    <Stack gap="xxl">
      <PageHeader title="Catálogo de trámites" />
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
        <Accordion variant="separated" chevronPosition="right">
          {entries.map((entry) => (
            <CatalogItem key={entry.id} entry={entry} />
          ))}
        </Accordion>
      )}
    </Stack>
  )
}
