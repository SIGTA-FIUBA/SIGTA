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
import { Link, useLocation, useOutletContext, useSearchParams } from 'react-router'
import { ForbiddenError, useProcedures } from '../api/procedures'
import type { components } from '../api/schema'

type Item = components['schemas']['ProcedureListItem']
type State = components['schemas']['ProcedureState']
type View = components['schemas']['ProcedureListView']
type Sort = components['schemas']['ProcedureListSort']
type User = components['schemas']['User']

const stateNames: Record<State, string> = {
  in_office: 'En la Secretaría',
  referred: 'En evaluación externa',
  with_applicant: 'Pendiente del titular',
}

const stateColors: Record<State, string> = {
  in_office: 'fiuba',
  referred: 'orange',
  with_applicant: 'gray',
}

const profileNames: Record<string, string> = {
  staff_operative: 'Perfil operativo',
  staff_approver: 'Perfil de aprobación',
}

const sortOptions = [
  { value: 'oldest', label: 'Mayor antigüedad' },
  { value: 'last_moved', label: 'Movimiento más reciente' },
]

const stateOptions = (Object.keys(stateNames) as State[]).map((state) => ({ value: state, label: stateNames[state] }))

type Filters = {
  view: View
  sort: Sort
  state?: State
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
    state: state !== null && state in stateNames ? (state as State) : undefined,
    procedureType: Number.isInteger(type) && type > 0 ? type : undefined,
    area: params.get('area') ?? undefined,
  }
}

// The API applies who can see what and what needs action. The page asks once
// for everything visible and filters, sorts and counts it here, which is only
// presentation: the list is small, since it holds procedures in progress.
function matches(item: Item, filters: Filters) {
  return (
    (filters.view === 'all' || item.actionable) &&
    (filters.state === undefined || item.state === filters.state) &&
    (filters.procedureType === undefined || item.procedure_type === filters.procedureType) &&
    (filters.area === undefined || item.area === filters.area)
  )
}

function compare(sort: Sort) {
  return (a: Item, b: Item) => {
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

function referredText(days: number) {
  return days === 1 ? '1 día sin retorno' : `${days} días sin retorno`
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

function StateCell({ item }: { item: Item }) {
  return (
    <Stack gap="xs" align="flex-start">
      <Badge variant="light" color={stateColors[item.state]}>
        {stateNames[item.state]}
      </Badge>
      {item.days_referred !== undefined && (
        <Text size="xs" c="orange.9" fw={600}>
          {referredText(item.days_referred)}
        </Text>
      )}
    </Stack>
  )
}

function ItemLink({ item, from }: { item: Item; from: string }) {
  return (
    <Anchor component={Link} to={`/procedures/${encodeURIComponent(item.id)}`} state={{ from }} fw={600}>
      {item.applicant_name}
    </Anchor>
  )
}

function ActionBadge({ item, show }: { item: Item; show: boolean }) {
  return show && item.actionable ? (
    <Badge variant="filled" size="sm">
      Acción requerida
    </Badge>
  ) : null
}

function ItemsTable({ items, from, showAction }: { items: Item[]; from: string; showAction: boolean }) {
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
                  <ItemLink item={item} from={from} />
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
                    {item.area_name}
                  </Text>
                </Stack>
              </Table.Td>
              <Table.Td>
                <StateCell item={item} />
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

function ItemCards({ items, from, showAction }: { items: Item[]; from: string; showAction: boolean }) {
  return (
    <Stack gap="sm" hiddenFrom="md">
      {items.map((item) => (
        <Paper key={item.id} component="article" withBorder p="md">
          <Stack gap="xs">
            <Group justify="space-between" align="flex-start" gap="xs">
              <Stack gap={0}>
                <ItemLink item={item} from={from} />
                <Text size="sm" c="dimmed">
                  {item.procedure_type_name}
                </Text>
              </Stack>
              <Badge variant="light" color={stateColors[item.state]}>
                {stateNames[item.state]}
              </Badge>
            </Group>
            <Stack gap={0}>
              <Group gap="xs">
                <Text fw={500}>{item.step_name}</Text>
                <ActionBadge item={item} show={showAction} />
              </Group>
              <Text size="sm" c="dimmed">
                {item.area_name}
              </Text>
            </Stack>
            <Group justify="space-between">
              <Text size="sm" c="dimmed">
                Iniciado hace {ageText(item.age_days).toLowerCase()}
              </Text>
              <Text size="sm" c={item.days_referred !== undefined ? 'orange.9' : 'dimmed'}>
                {item.days_referred !== undefined
                  ? referredText(item.days_referred)
                  : `Último movimiento: ${movedText(item.last_moved_at).toLowerCase()}`}
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
  const user = useOutletContext<User | null | undefined>()
  const filters = readFilters(params)

  const list = useProcedures()

  function update(changes: Record<string, string | null>) {
    const next = new URLSearchParams(params)
    for (const [name, value] of Object.entries(changes)) {
      if (value === null || value === '') {
        next.delete(name)
      } else {
        next.set(name, value)
      }
    }
    setParams(next, { replace: true })
  }

  const all = list.data?.items ?? []
  const count = (state: State) => all.filter((item) => item.state === state).length
  const actionable = all.filter((item) => item.actionable).length
  const shown = all.filter((item) => matches(item, filters)).sort(compare(filters.sort))
  const filtered = filters.procedureType !== undefined || filters.state !== undefined || filters.area !== undefined
  const from = `${location.pathname}${location.search}`

  const stateTile = (state: State, hint: string): Tile => ({
    key: state,
    label: stateNames[state],
    count: count(state),
    hint,
    selected: filters.view === 'all' && filters.state === state,
    onSelect: () => update({ view: 'all', state: filters.state === state ? null : state }),
  })
  const tiles: Tile[] = [
    {
      key: 'mine',
      label: 'Acción requerida',
      count: actionable,
      hint: 'Instancias asignadas a este perfil',
      selected: filters.view === 'mine' && !filtered,
      onSelect: () => update({ view: null, state: null, type: null, area: null }),
    },
    stateTile('in_office', 'En áreas de la Secretaría'),
    stateTile('referred', 'Derivados a evaluadores externos'),
    stateTile('with_applicant', 'Documentación o datos solicitados'),
  ]

  return (
    <Stack gap="lg">
      <Stack gap={0}>
        <Title order={1}>Bandeja de trámites</Title>
        {user != null && profileNames[user.role] !== undefined && <Text c="dimmed">{profileNames[user.role]}</Text>}
      </Stack>

      {list.isPending && <Loader aria-label="Cargando los trámites" />}
      {list.isError && list.error instanceof ForbiddenError && (
        <Alert role="alert">La bandeja de trámites está disponible solo para el personal de la Secretaría.</Alert>
      )}
      {list.isError && !(list.error instanceof ForbiddenError) && (
        <Alert color="red" role="alert">
          No se pudo cargar la bandeja de trámites. Volvé a intentarlo más tarde.
        </Alert>
      )}

      {list.isSuccess && (
        <>
          <SimpleGrid cols={{ base: 2, md: 4 }} spacing="sm">
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
                  onChange={(value) => update({ view: value === 'mine' ? null : value })}
                  data={[
                    { value: 'mine', label: `Asignados (${actionable})` },
                    { value: 'all', label: `Todos (${all.length})` },
                  ]}
                />
                <Group align="end" gap="sm">
                  <Select
                    label="Tipo de trámite"
                    placeholder="Todos"
                    clearable
                    data={list.data.procedure_types.map((option) => ({ value: option.id, label: option.name }))}
                    value={filters.procedureType === undefined ? null : String(filters.procedureType)}
                    onChange={(value) => update({ type: value })}
                  />
                  <Select
                    label="Estado"
                    placeholder="Todos"
                    clearable
                    data={stateOptions}
                    value={filters.state ?? null}
                    onChange={(value) => update({ state: value })}
                  />
                  <Select
                    label="Responsable"
                    placeholder="Todos"
                    clearable
                    data={list.data.areas.map((option) => ({ value: option.id, label: option.name }))}
                    value={filters.area ?? null}
                    onChange={(value) => update({ area: value })}
                  />
                  <Select
                    label="Ordenar por"
                    allowDeselect={false}
                    data={sortOptions}
                    value={filters.sort}
                    onChange={(value) => update({ sort: value === 'oldest' ? null : value })}
                  />
                </Group>
              </Group>

              {shown.length === 0 ? (
                <Text c="dimmed" p="md">
                  {filtered
                    ? 'No hay trámites que coincidan con los filtros.'
                    : filters.view === 'mine'
                      ? 'No hay trámites que requieran acción de este perfil.'
                      : 'No hay trámites en curso.'}
                </Text>
              ) : (
                <>
                  <ItemsTable items={shown} from={from} showAction={filters.view === 'all'} />
                  <ItemCards items={shown} from={from} showAction={filters.view === 'all'} />
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
