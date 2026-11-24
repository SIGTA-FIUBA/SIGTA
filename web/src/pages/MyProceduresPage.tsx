import { Alert, Skeleton, Stack, Text, TextInput, Title } from '@mantine/core'
import { useState, type ReactNode } from 'react'
import { useOutletContext } from 'react-router'
import { useMyProcedures } from '../api/procedures'
import type { components } from '../api/schema'
import { DocumentFile } from '../components/DocumentFile/DocumentFile'
import { EmptyState } from '../components/EmptyState/EmptyState'
import { Icon } from '../components/Icon/Icon'
import { MyProcedureRows } from '../components/MyProcedureRows/MyProcedureRows'
import { byLastMove, closedWithoutResolution } from '../content/procedures'
import { normalize, searchWords } from '../content/search'
import { PageHeader } from '../components/PageHeader/PageHeader'

type MyProcedure = components['schemas']['MyProcedure']
type User = components['schemas']['User']

// ClosedResolutions goes under the rows of the closed procedures, apart from
// them because each row is a link as a whole: for each one, the resolution to
// see or download, or that it closed without one.
function ClosedResolutions({ procedures }: { procedures: MyProcedure[] }) {
  return (
    <Stack gap="lg">
      {procedures.map((procedure) => (
        <Stack key={procedure.id} gap="xs">
          <Text fw={600}>{procedure.procedure_type_name}</Text>
          {procedure.resolutions.length === 0 ? (
            <Text size="sm" c="gray.7">
              {closedWithoutResolution}
            </Text>
          ) : (
            procedure.resolutions.map((resolution) => (
              <Stack key={resolution.id} gap="xxs">
                <Text size="sm" fw={500}>
                  {resolution.document_name}
                </Text>
                <DocumentFile
                  procedureId={procedure.id}
                  file={resolution}
                  procedureName={procedure.procedure_type_name}
                />
              </Stack>
            ))
          )}
        </Stack>
      ))}
    </Stack>
  )
}

// Section is a group of the page: its title and its rows.
function Section({ id, title, children }: { id: string; title: string; children: ReactNode }) {
  return (
    <Stack component="section" gap="md" aria-labelledby={id}>
      <Title order={2} id={id}>
        {title}
      </Title>
      {children}
    </Stack>
  )
}

// MyProceduresPage lists every procedure the person started, staff or not:
// in progress, by their latest move, and closed, with the resolution of
// each. A search by name narrows both.
export function MyProceduresPage() {
  // An applicant has their procedures in progress on the home page, which
  // calls this page its history.
  const user = useOutletContext<User | null | undefined>()
  const myProcedures = useMyProcedures()
  const [search, setSearch] = useState('')
  const words = searchWords(search)
  const found = (myProcedures.data ?? [])
    .filter((procedure) => words.every((word) => normalize(procedure.procedure_type_name).includes(word)))
    .sort(byLastMove)
  const inProgress = found.filter((procedure) => procedure.state !== 'closed')
  const closed = found.filter((procedure) => procedure.state === 'closed')

  return (
    <Stack gap="xxl">
      <Stack gap="lg" align="flex-start">
        <PageHeader title={user?.staff === true ? 'Mis trámites' : 'Historial'} />
        <TextInput
          aria-label="Buscar por nombre"
          placeholder="Buscar por nombre"
          leftSection={<Icon name="search" size={16} />}
          value={search}
          onChange={(event) => setSearch(event.currentTarget.value)}
        />
      </Stack>
      {myProcedures.isPending && (
        <Stack gap="xs" aria-label="Cargando los trámites">
          {Array.from({ length: 3 }, (_, index) => (
            <Skeleton key={index} h="var(--sigta-row-height)" />
          ))}
        </Stack>
      )}
      {myProcedures.isError && (
        <Alert color="red" role="alert">
          No se pudieron cargar los trámites. Volvé a intentarlo más tarde.
        </Alert>
      )}
      {myProcedures.isSuccess && myProcedures.data.length === 0 && (
        <EmptyState title="Sin trámites" text="Los trámites se inician desde el catálogo." />
      )}
      {myProcedures.isSuccess && myProcedures.data.length > 0 && found.length === 0 && (
        <Text c="dimmed">Ningún trámite coincide con “{search}”.</Text>
      )}
      {inProgress.length > 0 && (
        <Section id="in-progress" title="En curso">
          <MyProcedureRows label="Trámites en curso" procedures={inProgress} from="/my-procedures" />
        </Section>
      )}
      {closed.length > 0 && (
        <Section id="closed" title="Cerrados">
          <Stack gap="xl">
            <MyProcedureRows label="Trámites cerrados" procedures={closed} from="/my-procedures" />
            <ClosedResolutions procedures={closed} />
          </Stack>
        </Section>
      )}
    </Stack>
  )
}
