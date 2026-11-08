import { Alert, Anchor, Badge, Group, Loader, Paper, SegmentedControl, Select, Stack, Text, Title } from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import { Link, useLocation, useSearchParams } from 'react-router'
import { api } from '../api/client'
import type { components } from '../api/schema'

type Item = components['schemas']['ProcedureListItem']
type State = components['schemas']['ProcedureState']
type View = components['schemas']['ProcedureListView']
type Sort = components['schemas']['ProcedureListSort']

const stateNames: Record<State, string> = {
  in_office: 'En la Secretaría',
  referred: 'Derivado',
  with_applicant: 'Espera al titular',
}

const stateColors: Record<State, string> = {
  in_office: 'blue',
  referred: 'yellow',
  with_applicant: 'gray',
}

const sortOptions = [
  { value: 'oldest', label: 'Más antiguos primero' },
  { value: 'last_moved', label: 'Movidos más recientemente' },
]

const stateOptions = (Object.keys(stateNames) as State[]).map((state) => ({ value: state, label: stateNames[state] }))

// The filters live in the URL, so they survive going to a procedure and back,
// and a filtered list can be shared as a link.
function readFilters(params: URLSearchParams) {
  const view: View = params.get('view') === 'all' ? 'all' : 'mine'
  const sort: Sort = params.get('sort') === 'last_moved' ? 'last_moved' : 'oldest'
  const state = params.get('state')
  const type = Number(params.get('type'))
  return {
    view,
    sort,
    state: state !== null && state in stateNames ? (state as State) : undefined,
    procedureType: Number.isInteger(type) && type > 0 ? type : undefined,
    area: params.get('area') ?? undefined,
  }
}

class ForbiddenError extends Error {}

function daysText(days: number) {
  if (days === 0) {
    return 'hoy'
  }
  return days === 1 ? 'hace 1 día' : `hace ${days} días`
}

function ItemCard({ item, showActionable, from }: { item: Item; showActionable: boolean; from: string }) {
  return (
    <Paper component="article" withBorder p="md">
      <Stack gap="xs">
        <Group justify="space-between" wrap="wrap" gap="xs">
          <Anchor component={Link} to={`/procedures/${encodeURIComponent(item.id)}`} state={{ from }} fw={600}>
            {item.procedure_type_name}: {item.applicant_name}
          </Anchor>
          <Group gap="xs">
            {showActionable && item.actionable && <Badge variant="filled">Te toca</Badge>}
            <Badge variant="light" color={stateColors[item.state]}>
              {stateNames[item.state]}
            </Badge>
          </Group>
        </Group>
        <Text>
          Paso pendiente: {item.step_name}, a cargo de {item.area_name}.
        </Text>
        <Text size="sm" c="dimmed">
          Empezó {daysText(item.age_days)}.
          {item.days_referred !== undefined && ` Derivado ${daysText(item.days_referred)}, sin retorno.`}
        </Text>
      </Stack>
    </Paper>
  )
}

export function InboxPage() {
  const [params, setParams] = useSearchParams()
  const location = useLocation()
  const filters = readFilters(params)

  const list = useQuery({
    queryKey: ['procedures', filters],
    queryFn: async () => {
      const { data, response } = await api.GET('/procedures', {
        params: {
          query: {
            view: filters.view,
            sort: filters.sort,
            state: filters.state,
            procedure_type: filters.procedureType,
            area: filters.area,
          },
        },
      })
      if (response.status === 403) {
        throw new ForbiddenError()
      }
      if (data === undefined) {
        throw new Error(`GET /procedures answered ${response.status}`)
      }
      return data
    },
    placeholderData: (previous) => previous,
  })

  function setFilter(name: string, value: string | null) {
    const next = new URLSearchParams(params)
    if (value === null || value === '') {
      next.delete(name)
    } else {
      next.set(name, value)
    }
    setParams(next, { replace: true })
  }

  const filtered = filters.procedureType !== undefined || filters.state !== undefined || filters.area !== undefined
  const from = `${location.pathname}${location.search}`
  const items = list.data?.items ?? []

  return (
    <Stack gap="lg">
      <Title order={1}>Bandeja</Title>
      <SegmentedControl
        aria-label="Qué trámites ver"
        value={filters.view}
        onChange={(value) => setFilter('view', value === 'mine' ? null : value)}
        data={[
          { value: 'mine', label: 'Para mí' },
          { value: 'all', label: 'Todos' },
        ]}
      />
      <Group align="end" gap="sm" wrap="wrap">
        <Select
          label="Tipo de trámite"
          placeholder="Todos"
          clearable
          data={(list.data?.procedure_types ?? []).map((option) => ({ value: option.id, label: option.name }))}
          value={filters.procedureType === undefined ? null : String(filters.procedureType)}
          onChange={(value) => setFilter('type', value)}
        />
        <Select
          label="Estado"
          placeholder="Todos"
          clearable
          data={stateOptions}
          value={filters.state ?? null}
          onChange={(value) => setFilter('state', value)}
        />
        <Select
          label="Responsable"
          placeholder="Todos"
          clearable
          data={(list.data?.areas ?? []).map((option) => ({ value: option.id, label: option.name }))}
          value={filters.area ?? null}
          onChange={(value) => setFilter('area', value)}
        />
        <Select
          label="Ordenar por"
          allowDeselect={false}
          data={sortOptions}
          value={filters.sort}
          onChange={(value) => setFilter('sort', value === 'oldest' ? null : value)}
        />
      </Group>

      {list.isPending && <Loader aria-label="Cargando los trámites" />}
      {list.isError && list.error instanceof ForbiddenError && (
        <Alert role="alert">La bandeja es solo para el personal de la Secretaría.</Alert>
      )}
      {list.isError && !(list.error instanceof ForbiddenError) && (
        <Alert color="red" role="alert">
          No pudimos cargar la bandeja. Volvé a intentarlo más tarde.
        </Alert>
      )}
      {list.isSuccess && items.length === 0 && (
        <Text c="dimmed">
          {filtered
            ? 'No hay trámites con estos filtros.'
            : filters.view === 'mine'
              ? 'No tenés trámites pendientes.'
              : 'No hay trámites en curso.'}
        </Text>
      )}
      {list.isSuccess && items.length > 0 && (
        <Stack gap="sm">
          <Text size="sm" c="dimmed" aria-live="polite">
            {items.length === 1 ? '1 trámite' : `${items.length} trámites`}
          </Text>
          {items.map((item) => (
            <ItemCard key={item.id} item={item} showActionable={filters.view === 'all'} from={from} />
          ))}
        </Stack>
      )}
    </Stack>
  )
}
