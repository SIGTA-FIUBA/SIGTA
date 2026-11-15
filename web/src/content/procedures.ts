import type { components } from '../api/schema'
import type { StatusTone } from '../components/StatusBadge/StatusBadge'

type State = components['schemas']['ProcedureState']

// The states of a procedure as the Secretaría names them, with the tone of
// their badge.
export const procedureStates: Record<State, { tone: StatusTone; label: string }> = {
  in_office: { tone: 'office', label: 'En la Secretaría' },
  referred: { tone: 'external', label: 'En evaluación externa' },
  with_applicant: { tone: 'applicant', label: 'Pendiente del titular' },
}

// movedText says when something happened, counted in days.
export function movedText(timestamp: string) {
  const days = Math.max(0, Math.floor((Date.now() - Date.parse(timestamp)) / 86_400_000))
  if (days === 0) {
    return 'Hoy'
  }
  return days === 1 ? 'Hace 1 día' : `Hace ${days} días`
}
