import type { components } from '../api/schema'

type ProcedureTypeDraft = components['schemas']['ProcedureTypeDraft']

// draftStatus names where the draft of a procedure type stands, with the
// color of its badge. A rejected draft is being edited again, but says so
// until it is proposed again.
export function draftStatus(draft: Pick<ProcedureTypeDraft, 'status' | 'rejection'>): { label: string; color: string } {
  if (draft.status === 'proposed') {
    return { label: 'Propuesto', color: 'fiuba' }
  }
  return draft.rejection ? { label: 'Rechazado', color: 'orange' } : { label: 'En edición', color: 'gray' }
}
