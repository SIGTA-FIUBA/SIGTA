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
