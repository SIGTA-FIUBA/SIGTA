import type { components } from '../api/schema'
import { typeName } from './documentFormats'

type DefinitionProblem = components['schemas']['DefinitionProblem']

// The part of the editor where a problem is fixed: the route of the procedure
// type, its data and documents, or what describes it to the applicant.
export type ProblemSection = 'steps' | 'requirements' | 'general'

// ProblemPlace says where a problem goes, from its path: the step, field or
// document it is about, or the section when it is about none of them.
export type ProblemPlace = { section: ProblemSection; step?: string; field?: string; document?: string }

export function problemPlace(problem: DefinitionProblem): ProblemPlace {
  const step = problem.path.find((element) => element.kind === 'step')?.id
  if (step !== undefined) {
    return { section: 'steps', step }
  }
  const field = problem.path.find((element) => element.kind === 'field')?.id
  if (field !== undefined) {
    return { section: 'requirements', field }
  }
  const document = problem.path.find((element) => element.kind === 'document')?.id
  if (document !== undefined) {
    return { section: 'requirements', document }
  }
  if (problem.code === 'no_steps') {
    return { section: 'steps' }
  }
  return { section: 'general' }
}

// Names turns the identifiers that a problem carries into the names people
// gave them; an identifier without a name shows as it is.
export type Names = (id: string) => string

const kindNames: Record<string, string> = {
  definition: 'El tipo de trámite',
  field: 'El dato',
  option: 'La opción',
  document: 'El documento',
  step: 'El paso',
  outcome: 'La salida',
  deadline: 'El plazo',
  admissibility: 'La condición',
  area: 'El área',
  permission: 'El permiso',
}

// problemText says in Spanish what is wrong, about the innermost element of
// the path, which the place already shows. The server's text is in English
// and only for logs.
export function problemText(problem: DefinitionProblem, names: Names): string {
  const last = problem.path.at(-1)
  const subject = kindNames[last?.kind ?? 'definition'] ?? 'Un elemento'
  const value = problem.value ?? ''
  const texts: Record<string, string> = {
    no_id: `${subject} no tiene identificador.`,
    invalid_id: `${subject} tiene un identificador que no sirve: «${value}».`,
    no_name: `${subject} no tiene nombre.`,
    name_with_spaces: `El nombre «${value}» empieza o termina con espacios.`,
    declared_twice: `${subject} está dos veces.`,
    repeated_name: `El nombre «${value}» está repetido.`,
    name_in_use: `El nombre «${value}» ya lo usa el área «${names(problem.other ?? '')}».`,
    no_steps: 'El tipo de trámite no tiene pasos.',
    unused_field: 'Ningún paso pide este dato.',
    unused_document: 'Ningún paso pide este documento.',
    unreachable_step: `No se llega a este paso desde «${names(value)}».`,
    step_never_closes: 'Desde este paso nunca se llega a cerrar el trámite.',
    outcome_loops: 'La salida vuelve al mismo paso.',
    send_back_forward: 'La salida es una devolución, pero no lleva a un paso anterior.',
    resolution_from_applicant: 'Es la resolución del trámite, que carga el personal, pero la pide un paso del titular.',
    unknown_area: `El área «${value}» no existe.`,
    unknown_role: 'El paso tiene que ser del titular o del personal.',
    applicant_with_area: 'Un paso del titular no tiene área.',
    applicant_with_permissions: 'Un paso del titular no exige permisos.',
    applicant_action: 'El titular solo carga datos o documentos.',
    no_permissions: 'Un paso del personal exige al menos un permiso.',
    unknown_permission: `El área «${names(problem.other ?? '')}» no tiene el permiso «${value}».`,
    repeated_permission: `El permiso «${value}» está repetido.`,
    unknown_action: 'La acción del paso no existe.',
    unknown_field: `El dato «${value}» no existe.`,
    unknown_document: `El documento «${value}» no existe.`,
    repeated_field: `El dato «${names(value)}» está repetido.`,
    repeated_document: `El documento «${names(value)}» está repetido.`,
    enter_data_no_fields: 'Un paso que carga datos pide al menos un dato.',
    enter_data_with_documents: 'Un paso que carga datos no pide documentos.',
    upload_no_documents: 'Un paso que carga documentos pide al menos un documento.',
    upload_with_fields: 'Un paso que carga documentos no pide datos.',
    confirm_with_documents: 'Un paso que confirma un hecho externo no pide documentos.',
    decide_with_next: 'Una decisión sigue por sus salidas, no por un paso siguiente.',
    decide_few_outcomes: 'Una decisión necesita al menos dos salidas.',
    outcomes_outside_decide: 'Solo una decisión tiene salidas.',
    unknown_next_step: `El paso «${value}» no existe.`,
    unknown_field_type: 'El tipo del dato no existe.',
    options_outside_option: 'Solo un dato de opciones tiene opciones.',
    option_field_no_options: 'Un dato de opciones necesita al menos una opción.',
    no_description: 'Falta la descripción del tipo de trámite.',
    blank_text: `${subject} no tiene texto.`,
    no_accepted_types: 'El documento no acepta ningún formato.',
    invalid_accepted_type: `«${value}» no es un formato.`,
    undetected_accepted_type: `El formato ${typeName(value)} no se reconoce en un archivo.`,
    repeated_accepted_type: `El formato ${typeName(value)} está repetido.`,
    no_max_size: 'El documento no tiene tamaño máximo.',
  }
  return texts[problem.code] ?? 'Hay un problema en la definición.'
}
