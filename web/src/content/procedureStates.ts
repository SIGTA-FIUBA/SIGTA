import type { components } from '../api/schema'

type State = components['schemas']['ProcedureState']

export const stateNames: Record<State, string> = {
  in_office: 'En un área',
  referred: 'Derivado',
  with_applicant: 'Pendiente del titular',
  closed: 'Cerrado',
}

export const stateColors: Record<State, string> = {
  in_office: 'fiuba',
  referred: 'orange',
  with_applicant: 'gray',
  closed: 'gray',
}

// The home page and the detail say the same when a procedure closed without
// a document that resolves it.
export const closedWithoutResolution = 'El trámite se cerró sin un documento resolutivo adjunto.'
