import { Alert, Grid, SegmentedControl, Select, SimpleGrid, Skeleton, Stack, Text, Title } from '@mantine/core'
import { useLocation, useSearchParams } from 'react-router'
import { ForbiddenError, useInboxProcedures } from '../api/procedures'
import type { components } from '../api/schema'
import { Figure } from '../components/Figure/Figure'
import { Figures } from '../components/Figures/Figures'
import { ProcedureRow } from '../components/ProcedureRow/ProcedureRow'
import { ProcedureRows } from '../components/ProcedureRows/ProcedureRows'
import { movedText, procedurePath, stateNames, standing } from '../content/procedures'

type ProcedureListItem = components['schemas']['ProcedureListItem']
type ProcedureState = components['schemas']['ProcedureState']
type ProcedureListView = components['schemas']['ProcedureListView']
type ProcedureListSort = components['schemas']['ProcedureListSort']

const inboxStates: ProcedureState[] = ['in_office', 'referred', 'with_applicant']

const sortOptions = [
  { value: 'oldest', label: 'Mayor antigüedad' },
  { value: 'last_moved', label: 'Movimiento más reciente' },
]

const stateOptions = inboxStates.map((state) => ({ value: state, label: stateNames[state] }))

type Filters = {
  view: ProcedureListView
  sort: ProcedureListSort
  state?: ProcedureState
  procedureType?: number
  area?: string
}

// The filters live in the URL, so they survive going to a procedure and back,
// and a filtered list can be shared as a link.
function readFilters(params: URLSearchParams): Filters {
  const state = params.get('state')
  const type = Number(params.get('type'))
  return {
    view: params.get('view') === 'all' ? 'all' : 'mine',
    sort: params.get('sort') === 'last_moved' ? 'last_moved' : 'oldest',
    state: inboxStates.find((known) => known === state),
    procedureType: Number.isInteger(type) && type > 0 ? type : undefined,
    area: params.get('area') ?? undefined,
  }
}

// The API applies who can see what and what needs action. The page asks once
// for everything visible and filters, sorts and counts it here, which is only
// presentation: the list is small, since it holds procedures in progress.
function matchesFilters(item: ProcedureListItem, filters: Filters) {
  return (
    (filters.view === 'all' || item.actionable) &&
    (filters.state === undefined || item.state === filters.state) &&
    (filters.procedureType === undefined || item.procedure_type === filters.procedureType) &&
    (filters.area === undefined || item.area?.id === filters.area)
  )
}

function compare(sort: ProcedureListSort) {
  return (a: ProcedureListItem, b: ProcedureListItem) => {
    const byDate =
      sort === 'last_moved'
        ? Date.parse(b.last_moved_at) - Date.parse(a.last_moved_at)
        : Date.parse(a.started_at) - Date.parse(b.started_at)
    return byDate !== 0 ? byDate : a.id.localeCompare(b.id)
  }
}

// ageText says how long a procedure has been in progress.
function ageText(days: number) {
  if (days === 0) {
    return 'Iniciado hoy'
  }
  return days === 1 ? 'Iniciado hace 1 día' : `Iniciado hace ${days} días`
}

const columns = ['Trámite', 'Paso', 'A cargo de', 'Último movimiento'] as const

// ProcedureList is the list of the inbox, in the rows of the home page: the
// type and the applicant, the open step, who has it and the last news with
// the age. In the view of everything, what needs action of this account says
// so. The rows travel to their new place when the order or a filter changes.
function ProcedureList({ items, from, showAction }: { items: ProcedureListItem[]; from: string; showAction: boolean }) {
  return (
    <ProcedureRows label="Trámites de la bandeja" columns={columns}>
      {items.map((item) => (
        <ProcedureRow
          key={item.id}
          to={procedurePath(item.id)}
          state={{ from }}
          type={item.procedure_type_name}
          applicant={item.applicant.name || 'Titular sin nombre registrado'}
          note={showAction && item.actionable ? 'Acción requerida' : undefined}
          step={{ name: item.step_name }}
          standing={standing(item)}
          date={movedText(item.last_moved_at)}
          age={ageText(item.age_days)}
        />
      ))}
    </ProcedureRows>
  )
}

export function InboxPage() {
  const [params, setParams] = useSearchParams()
  const location = useLocation()
  const filters = readFilters(params)

  const procedures = useInboxProcedures()

  function updateParams(changes: Record<string, string | null>) {
    const next = new URLSearchParams(params)
    for (const [name, value] of Object.entries(changes)) {
      if (value === null) {
        next.delete(name)
      } else {
        next.set(name, value)
      }
    }
    setParams(next, { replace: true })
  }

  const all = procedures.data?.items ?? []
  const countByState = (state: ProcedureState) => all.filter((item) => item.state === state).length
  const actionableCount = all.filter((item) => item.actionable).length
  const shown = all.filter((item) => matchesFilters(item, filters)).sort(compare(filters.sort))
  const hasFilters = filters.procedureType !== undefined || filters.state !== undefined || filters.area !== undefined
  const from = `${location.pathname}${location.search}`

  // The figures of the inbox, which filter the list: what needs action of
  // this account and each state, counted over everything visible.
  const stateTile = (state: ProcedureState) => ({
    key: state,
    label: stateNames[state],
    value: countByState(state),
    selected: filters.view === 'all' && filters.state === state,
    onSelect: () => updateParams({ view: 'all', state: filters.state === state ? null : state }),
  })
  const tiles = [
    {
      key: 'mine',
      label: 'Acción requerida',
      value: actionableCount,
      selected: filters.view === 'mine' && !hasFilters,
      onSelect: () => updateParams({ view: null, state: null, type: null, area: null }),
    },
    stateTile('in_office'),
    stateTile('referred'),
    stateTile('with_applicant'),
  ]

  return (
    <Stack gap="xl">
      <Stack gap="xs">
        <Title order={1} variant="display" c="fiuba.9">
          Bandeja de trámites
        </Title>
        <Text c="gray.7">Los trámites en curso que ve esta cuenta, con el paso en que está cada uno.</Text>
      </Stack>

      {procedures.isPending && (
        <Stack gap="sm" aria-label="Cargando los trámites">
          <Skeleton h="var(--sigta-figures-height)" />
          <Skeleton h="var(--sigta-row-height)" />
          <Skeleton h="var(--sigta-row-height)" />
        </Stack>
      )}
      {procedures.isError && procedures.error instanceof ForbiddenError && (
        <Alert role="alert">La bandeja de trámites está disponible solo para el personal.</Alert>
      )}
      {procedures.isError && !(procedures.error instanceof ForbiddenError) && (
        <Alert color="red" role="alert">
          No se pudo cargar la bandeja de trámites. Volvé a intentarlo más tarde.
        </Alert>
      )}

      {procedures.isSuccess && (
        <>
          <Figures label="Resumen por estado">
            {tiles.map(({ key, ...tile }) => (
              <Figure key={key} {...tile} />
            ))}
          </Figures>

          {/* On a wide screen the filters take the first quarter of the
              grid, as the titles of the home page, and the list the rest. */}
          <Grid gap={{ base: 'lg', lg: 'xl' }}>
            <Grid.Col span={{ base: 12, lg: 3 }}>
              <Stack gap="md" component="search" aria-label="Filtros">
                <SegmentedControl
                  aria-label="Trámites a mostrar"
                  fullWidth
                  value={filters.view}
                  onChange={(value) => updateParams({ view: value === 'mine' ? null : value })}
                  data={[
                    { value: 'mine', label: `Asignados (${actionableCount})` },
                    { value: 'all', label: `Todos (${all.length})` },
                  ]}
                />
                <SimpleGrid cols={{ base: 2, lg: 1 }} spacing="md">
                  <Select
                    label="Tipo de trámite"
                    placeholder="Todos"
                    clearable
                    data={procedures.data.procedure_types.map((option) => ({ value: option.id, label: option.name }))}
                    value={filters.procedureType === undefined ? null : String(filters.procedureType)}
                    onChange={(value) => updateParams({ type: value })}
                  />
                  <Select
                    label="Estado"
                    placeholder="Todos"
                    clearable
                    data={stateOptions}
                    value={filters.state ?? null}
                    onChange={(value) => updateParams({ state: value })}
                  />
                  <Select
                    label="Responsable"
                    placeholder="Todos"
                    clearable
                    data={procedures.data.areas.map((option) => ({ value: option.id, label: option.name }))}
                    value={filters.area ?? null}
                    onChange={(value) => updateParams({ area: value })}
                  />
                  <Select
                    label="Ordenar por"
                    allowDeselect={false}
                    data={sortOptions}
                    value={filters.sort}
                    onChange={(value) => updateParams({ sort: value === 'oldest' ? null : value })}
                  />
                </SimpleGrid>
                <Text size="sm" c="gray.7" aria-live="polite">
                  {shown.length === 1 ? '1 trámite' : `${shown.length} trámites`}
                </Text>
              </Stack>
            </Grid.Col>
            <Grid.Col span={{ base: 12, lg: 9 }}>
              {shown.length === 0 ? (
                <Text c="gray.7">
                  {hasFilters
                    ? 'No hay trámites que coincidan con los filtros.'
                    : filters.view === 'mine'
                      ? 'No hay trámites que requieran acción de esta cuenta.'
                      : 'No hay trámites en curso.'}
                </Text>
              ) : (
                <ProcedureList items={shown} from={from} showAction={filters.view === 'all'} />
              )}
            </Grid.Col>
          </Grid>
        </>
      )}
    </Stack>
  )
}
