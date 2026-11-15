import { Anchor, Button, Loader, Stack, Title } from '@mantine/core'
import { Link, useOutletContext } from 'react-router'
import { useProcedures } from '../api/procedures'
import type { components } from '../api/schema'
import { EmptyState } from '../components/EmptyState/EmptyState'
import { PageHeader } from '../components/PageHeader/PageHeader'
import { ProcedureRow, ProcedureRows } from '../components/ProcedureRow/ProcedureRow'
import type { StatusTone } from '../components/StatusBadge/StatusBadge'
import { StepList } from '../components/StepList/StepList'
import { applicantSteps } from '../content/platform'

type User = components['schemas']['User']
type Item = components['schemas']['ProcedureListItem']
type State = components['schemas']['ProcedureState']

const states: Record<State, { tone: StatusTone; label: string }> = {
  in_office: { tone: 'office', label: 'En la Secretaría' },
  referred: { tone: 'external', label: 'En evaluación externa' },
  with_applicant: { tone: 'applicant', label: 'Pendiente del titular' },
}

// The home page shows the latest procedures that need action, no more.
const latest = 5

function movedText(timestamp: string) {
  const days = Math.max(0, Math.floor((Date.now() - Date.parse(timestamp)) / 86_400_000))
  if (days === 0) {
    return 'Hoy'
  }
  return days === 1 ? 'Hace 1 día' : `Hace ${days} días`
}

// Pending lists, for the staff, the procedures that need action from the
// profile, the latest news first.
function Pending() {
  const procedures = useProcedures()
  if (procedures.isPending) {
    return <Loader size="sm" aria-label="Cargando los trámites" />
  }
  if (procedures.isError) {
    return <EmptyState title="No se pudieron cargar los trámites" text="Volvé a intentarlo en unos minutos." />
  }
  const items = procedures.data.items
    .filter((item) => item.actionable)
    .sort((a, b) => Date.parse(b.last_moved_at) - Date.parse(a.last_moved_at))
  if (items.length === 0) {
    return (
      <EmptyState
        title="No hay trámites pendientes"
        text="Cuando un trámite requiera acción del perfil, aparece acá."
      />
    )
  }
  return (
    <Stack gap="sm">
      <ProcedureRows label="Trámites que requieren acción">
        {items.slice(0, latest).map((item: Item) => (
          <ProcedureRow
            key={item.id}
            to={`/procedures/${encodeURIComponent(item.id)}`}
            state={{ from: '/' }}
            name={item.procedure_type_name}
            detail={item.applicant_name}
            status={states[item.state]}
            nextStep={item.step_name}
            date={movedText(item.last_moved_at)}
          />
        ))}
      </ProcedureRows>
      {items.length > latest && (
        <Anchor component={Link} to="/inbox" size="sm" w="fit-content">
          Ver los {items.length} trámites en la bandeja
        </Anchor>
      )}
    </Stack>
  )
}

// HomePage shows the person what to do next. The staff gets the inbox as the
// primary action and the procedures that need action; the applicant, the way
// to start a procedure and how it is done. What SIGTA is lives in "Acerca de
// SIGTA", at the foot of the menu.
export function HomePage() {
  // The layout passes the person with the session; outside it there is none.
  const user = useOutletContext<User | null | undefined>()
  const staff = user != null && user.role !== 'applicant'

  return (
    <Stack gap="xxl">
      <PageHeader
        title="Inicio"
        action={
          staff ? (
            <Button component={Link} to="/inbox">
              Ir a la bandeja
            </Button>
          ) : (
            <Button component={Link} to="/catalog">
              Iniciar un trámite
            </Button>
          )
        }
      />
      {staff && (
        <Stack component="section" aria-labelledby="pending" gap="md">
          <Title order={2} id="pending">
            Requieren acción
          </Title>
          <Pending />
        </Stack>
      )}
      {!staff && (
        <Stack component="section" aria-labelledby="how" gap="md">
          <Title order={2} id="how">
            Cómo se realiza un trámite
          </Title>
          <StepList steps={applicantSteps} />
        </Stack>
      )}
    </Stack>
  )
}
