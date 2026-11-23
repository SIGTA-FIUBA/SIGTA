import type { components } from '../api/schema'
import type { StatusTone } from '../components/StatusBadge/StatusBadge'

type State = components['schemas']['ProcedureState']
type Step = components['schemas']['ProcedureStep']

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

// The home page and the detail say the same when a procedure closed without
// a document that resolves it.
export const closedWithoutResolution = 'El trámite se cerró sin un documento resolutivo adjunto.'

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

// requirementsOf names each field and document the steps ask for, once each,
// in the order of the steps.
export function requirementsOf(steps: Step[]) {
  const fields = new Map<string, string>()
  const documents = new Map<string, string>()
  for (const step of steps) {
    for (const field of step.fields) {
      if (!fields.has(field.id)) {
        fields.set(field.id, field.name)
      }
    }
    for (const document of step.documents) {
      if (!documents.has(document.id)) {
        documents.set(document.id, document.name)
      }
    }
  }
  return { fields, documents }
}

// requestedNames writes what a correction asks for, by the names the steps
// give each field and document.
export function requestedNames(steps: Step[], requested: { fields: string[]; documents: string[] }) {
  const { fields, documents } = requirementsOf(steps)
  return [
    ...requested.fields.map((id) => fields.get(id) ?? id),
    ...requested.documents.map((id) => documents.get(id) ?? id),
  ].join(', ')
}

// byLastMove sorts procedures by their latest move, the newest first.
export function byLastMove(a: { last_moved_at: string }, b: { last_moved_at: string }) {
  return Date.parse(b.last_moved_at) - Date.parse(a.last_moved_at)
}
