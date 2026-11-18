import {
  Alert,
  Anchor,
  Badge,
  Group,
  Loader,
  Paper,
  SegmentedControl,
  Select,
  SimpleGrid,
  Stack,
  Table,
  Text,
  Title,
} from '@mantine/core'
import { Link, useLocation, useSearchParams } from 'react-router'
import { ForbiddenError, useProcedures } from '../api/procedures'
import type { components } from '../api/schema'
import { stateColors, stateNames } from '../procedureState'

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

function daysSince(timestamp: string) {
  return Math.max(0, Math.floor((Date.now() - Date.parse(timestamp)) / 86_400_000))
}

function ageText(days: number) {
  if (days === 0) {
    return 'Hoy'
  }
  return days === 1 ? '1 día' : `${days} días`
}

function movedText(timestamp: string) {
  const days = daysSince(timestamp)
  if (days === 0) {
    return 'Hoy'
  }
  return days === 1 ? 'Hace 1 día' : `Hace ${days} días`
}

type Tile = { key: string; label: string; count: number; hint: string; selected: boolean; onSelect: () => void }

function SummaryTile({ tile }: { tile: Tile }) {
  return (
    <Paper
      component="button"
      type="button"
      variant="interactive"
      onClick={tile.onSelect}
      aria-pressed={tile.selected}
      withBorder
      p="md"
      w="100%"
      h="100%"
      ta="left"
      bg={tile.selected ? 'fiuba.0' : 'white'}
    >
      <Stack gap="xs">
        <Text size="sm" fw={500} c="dimmed">
          {tile.label}
        </Text>
        <Text fz="xl" fw={700}>
          {tile.count}
        </Text>
        <Text size="xs" c="dimmed">
          {tile.hint}
        </Text>
      </Stack>
    </Paper>
  )
}

function StateBadge({ item }: { item: ProcedureListItem }) {
  return (
    <Badge variant="light" color={stateColors[item.state]}>
      {stateNames[item.state]}
    </Badge>
  )
}

// A step of the applicant has no area: the applicant is shown as the one in
// charge.
function responsibleText(item: ProcedureListItem) {
  return item.area?.name ?? 'Titular'
}

function ProcedureLink({ item, from }: { item: ProcedureListItem; from: string }) {
  return (
    <Anchor component={Link} to={`/procedures/${encodeURIComponent(item.id)}`} state={{ from }} fw={600}>
      {item.applicant.name || 'Titular sin nombre registrado'}
    </Anchor>
  )
}

function ActionBadge({ item, show }: { item: ProcedureListItem; show: boolean }) {
  return show && item.actionable ? (
    <Badge variant="filled" size="sm">
      Acción requerida
    </Badge>
  ) : null
}

function ProceduresTable({
  items,
  from,
  showAction,
}: {
  items: ProcedureListItem[]
  from: string
  showAction: boolean
}) {
  return (
    <Table.ScrollContainer minWidth={860} visibleFrom="md">
      <Table verticalSpacing="sm" highlightOnHover>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Trámite</Table.Th>
            <Table.Th>Instancia pendiente</Table.Th>
            <Table.Th>Estado</Table.Th>
            <Table.Th ta="right">Antigüedad</Table.Th>
            <Table.Th ta="right">Último movimiento</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {items.map((item) => (
            <Table.Tr key={item.id}>
              <Table.Td>
                <Stack gap={0}>
                  <ProcedureLink item={item} from={from} />
                  <Text size="sm" c="dimmed">
                    {item.procedure_type_name}
                  </Text>
                </Stack>
              </Table.Td>
              <Table.Td>
                <Stack gap={0}>
                  <Group gap="xs">
                    <Text fw={500}>{item.step_name}</Text>
                    <ActionBadge item={item} show={showAction} />
                  </Group>
                  <Text size="sm" c="dimmed">
                    {responsibleText(item)}
                  </Text>
                </Stack>
              </Table.Td>
              <Table.Td>
                <StateBadge item={item} />
              </Table.Td>
              <Table.Td ta="right">{ageText(item.age_days)}</Table.Td>
              <Table.Td ta="right">
                <Text size="sm" c="dimmed">
                  {movedText(item.last_moved_at)}
                </Text>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </Table.ScrollContainer>
  )
}

function ProcedureCards({
  items,
  from,
  showAction,
}: {
  items: ProcedureListItem[]
  from: string
  showAction: boolean
}) {
  return (
    <Stack gap="sm" hiddenFrom="md">
      {items.map((item) => (
        <Paper key={item.id} component="article" withBorder p="md">
          <Stack gap="xs">
            <Group justify="space-between" align="flex-start" gap="xs">
              <Stack gap={0}>
                <ProcedureLink item={item} from={from} />
                <Text size="sm" c="dimmed">
                  {item.procedure_type_name}
                </Text>
              </Stack>
              <StateBadge item={item} />
            </Group>
            <Stack gap={0}>
              <Group gap="xs">
                <Text fw={500}>{item.step_name}</Text>
                <ActionBadge item={item} show={showAction} />
              </Group>
              <Text size="sm" c="dimmed">
                {responsibleText(item)}
              </Text>
            </Stack>
            <Group justify="space-between">
              <Text size="sm" c="dimmed">
                Iniciado hace {ageText(item.age_days).toLowerCase()}
              </Text>
              <Text size="sm" c="dimmed">
                Último movimiento: {movedText(item.last_moved_at).toLowerCase()}
              </Text>
            </Group>
          </Stack>
        </Paper>
      ))}
    </Stack>
  )
}

export function InboxPage() {
  const [params, setParams] = useSearchParams()
  const location = useLocation()
  const filters = readFilters(params)

  const procedures = useProcedures()

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

  const stateTile = (state: ProcedureState, hint: string): Tile => ({
    key: state,
    label: stateNames[state],
    count: countByState(state),
    hint,
    selected: filters.view === 'all' && filters.state === state,
    onSelect: () => updateParams({ view: 'all', state: filters.state === state ? null : state }),
  })
  const tiles: Tile[] = [
    {
      key: 'mine',
      label: 'Acción requerida',
      count: actionableCount,
      hint: 'Pasos abiertos que exigen los permisos de esta cuenta',
      selected: filters.view === 'mine' && !hasFilters,
      onSelect: () => updateParams({ view: null, state: null, type: null, area: null }),
    },
    stateTile('in_office', 'En las áreas del personal'),
    stateTile('referred', 'En otra área, sin volver todavía al área de la que salió'),
    stateTile('with_applicant', 'Documentación o datos solicitados'),
  ]

  return (
    <Stack gap="lg">
      <Title order={1}>Bandeja de trámites</Title>

      {procedures.isPending && <Loader aria-label="Cargando los trámites" />}
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
          <SimpleGrid cols={{ base: 1, xs: 3 }} spacing="sm">
            {tiles.map((tile) => (
              <SummaryTile key={tile.key} tile={tile} />
            ))}
          </SimpleGrid>

          <Paper withBorder>
            <Stack gap={0}>
              <Group justify="space-between" align="end" gap="sm" p="md">
                <SegmentedControl
                  aria-label="Trámites a mostrar"
                  value={filters.view}
                  onChange={(value) => updateParams({ view: value === 'mine' ? null : value })}
                  data={[
                    { value: 'mine', label: `Asignados (${actionableCount})` },
                    { value: 'all', label: `Todos (${all.length})` },
                  ]}
                />
                <Group align="end" gap="sm">
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
                </Group>
              </Group>

              {shown.length === 0 ? (
                <Text c="dimmed" p="md">
                  {hasFilters
                    ? 'No hay trámites que coincidan con los filtros.'
                    : filters.view === 'mine'
                      ? 'No hay trámites que requieran acción de esta cuenta.'
                      : 'No hay trámites en curso.'}
                </Text>
              ) : (
                <>
                  <ProceduresTable items={shown} from={from} showAction={filters.view === 'all'} />
                  <ProcedureCards items={shown} from={from} showAction={filters.view === 'all'} />
                </>
              )}
              <Text size="sm" c="dimmed" p="md" aria-live="polite">
                {shown.length === 1 ? '1 trámite' : `${shown.length} trámites`}
              </Text>
            </Stack>
          </Paper>
        </>
      )}
    </Stack>
  )
}
