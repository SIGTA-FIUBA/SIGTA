import type { components } from '../api/schema'
import type { StatusTone } from '../components/StatusBadge/StatusBadge'

type State = components['schemas']['ProcedureState']

export const stateNames: Record<State, string> = {
  in_office: 'En un área',
  referred: 'Derivado',
  with_applicant: 'Pendiente del titular',
  closed: 'Cerrado',
}

// The states of a procedure with the tone of their badge. The names are the
// same as everywhere else.
export const procedureStates: Record<State, { tone: StatusTone; label: string }> = {
  in_office: { tone: 'office', label: stateNames.in_office },
  referred: { tone: 'external', label: stateNames.referred },
  with_applicant: { tone: 'applicant', label: stateNames.with_applicant },
  closed: { tone: 'closed', label: stateNames.closed },
}

// procedurePath is the page of a procedure.
export function procedurePath(id: string) {
  return `/procedures/${encodeURIComponent(id)}`
}

// movedText says when something happened, counted in days.
export function movedText(timestamp: string) {
  const days = Math.max(0, Math.floor((Date.now() - Date.parse(timestamp)) / 86_400_000))
  if (days === 0) {
    return 'Hoy'
  }
  return days === 1 ? 'Hace 1 día' : `Hace ${days} días`
}

type Item = components['schemas']['ProcedureListItem']

// standing says who acts now on a procedure of the inbox, for the third
// column of its row. One waiting on its applicant is in no area's hands, even
// if its open step belongs to one; one referred says so after the area.
export function standing(item: Item): { label?: string; text: string } {
  if (item.state === 'with_applicant' || !item.area) {
    return { text: procedureStates[item.state].label }
  }
  return {
    label: 'A cargo de',
    text: item.state === 'referred' ? `${item.area.name} (derivado)` : item.area.name,
  }
}
