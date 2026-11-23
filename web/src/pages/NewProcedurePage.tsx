import { Alert, Anchor, SimpleGrid, Skeleton, Stack, Text } from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router'
import { catalogQuery } from '../api/catalog'
import { EmptyState } from '../components/EmptyState/EmptyState'
import { PageHeader } from '../components/PageHeader/PageHeader'
import { Reveal } from '../components/Reveal/Reveal'
import { StartCard } from '../components/StartCard/StartCard'

// NewProcedurePage is the gallery of the procedures to start: every type of
// the catalog as a card, by name, in rows of three that close. What each one
// is in more detail is in the catalog.
export function NewProcedurePage() {
  const catalog = useQuery(catalogQuery)
  const entries = [...(catalog.data ?? [])].sort((a, b) => a.name.localeCompare(b.name, 'es'))
  return (
    <Stack gap="xl">
      <PageHeader title="Iniciar trámite" />
      {catalog.isError ? (
        <Alert color="red" role="alert">
          No se pudo cargar el catálogo. Volvé a intentarlo más tarde.
        </Alert>
      ) : catalog.isPending ? (
        <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }} aria-label="Cargando los trámites">
          {Array.from({ length: 3 }, (_, index) => (
            <Skeleton key={index} h="var(--sigta-figures-height)" />
          ))}
        </SimpleGrid>
      ) : entries.length === 0 ? (
        <EmptyState title="Sin trámites publicados" text="Cuando se publique un tipo de trámite, aparece acá." />
      ) : (
        <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }} spacing={{ base: 'xs', sm: 'md' }}>
          {entries.map((entry, index) => (
            <Reveal key={entry.id} index={index}>
              <StartCard to={`/catalog/${entry.id}/start`} name={entry.name} description={entry.description} />
            </Reveal>
          ))}
        </SimpleGrid>
      )}
      <Text size="sm" c="gray.7">
        Qué es cada trámite y quién puede pedirlo, en el{' '}
        <Anchor component={Link} to="/catalog" inherit fw={600}>
          catálogo de trámites
        </Anchor>
        .
      </Text>
    </Stack>
  )
}
