import type { components } from '../api/schema'

// Details for the preview of the page of a procedure, as staff who can act on
// a step that asks for documents: kind data, the realistic worst case, and the
// edge counts. Synthetic data only (ADR-005).

type Detail = components['schemas']['ProcedureDetail']
type Step = components['schemas']['ProcedureStep']
type FieldRequirement = components['schemas']['FieldRequirement']
type DocumentRequirement = components['schemas']['DocumentRequirement']
type Value = components['schemas']['ProcedureValue']
type Event = components['schemas']['ProcedureEvent']
type Attached = components['schemas']['ProcedureDocument']

const megabyte = 1024 * 1024
const applicant = { id: 'dev-applicant-01', name: 'Lucía Ejemplo' }
const staff = { id: 'dev-operative', name: 'Operativo de prueba' }
const at = '2026-11-16T12:00:00Z'

const field = (id: string, name: string, type: FieldRequirement['type']): FieldRequirement => ({
  id,
  name,
  required: true,
  type,
  options: [],
})

const document = (id: string, name: string, required = true): DocumentRequirement => ({
  id,
  name,
  required,
  accepted_types: ['application/pdf'],
  max_size: 10 * megabyte,
})

const text = (fieldId: string, value: string): Value => ({
  field: fieldId,
  step: 'request',
  text: value,
  entered_by: applicant,
  entered_at: at,
})

function detail(
  fields: FieldRequirement[],
  values: Value[],
  documents: DocumentRequirement[],
  history: Event[],
  files: Attached[] = [],
): Detail {
  const request: Step = {
    id: 'request',
    name: 'Solicitud',
    action: 'enter_data',
    role: 'applicant',
    fields,
    documents: [],
    outcomes: [],
    next: 'communicate',
    values,
    files: [],
  }
  const communicate: Step = {
    id: 'communicate',
    name: 'Comunicación del dictamen',
    action: 'upload_document',
    role: 'staff',
    area: { id: 'office', name: 'Dirección de Alumnos' },
    fields: [],
    documents,
    outcomes: [],
    values: [],
    files,
  }
  return {
    id: 'dev-09',
    procedure_type: 3,
    version: 1,
    name: 'Trámite de ejemplo con dictamen',
    state: 'in_office',
    applicant,
    open_step: 'communicate',
    can_act: true,
    steps: [request, communicate],
    history,
  }
}

const started: Event = {
  step: 'request',
  by: applicant,
  at,
  observation: '',
  declared: false,
  send_back: false,
  returned: false,
  next: 'communicate',
}

const attached = (id: string, requirement: string, fileName: string): Event => ({
  step: 'communicate',
  by: staff,
  at: '2026-11-17T15:00:00Z',
  observation: '',
  declared: false,
  send_back: false,
  returned: false,
  document: { id, document: requirement, file_name: fileName },
})

const file = (id: string, requirement: string, fileName: string): Attached => ({
  id,
  document: requirement,
  step: 'communicate',
  file_name: fileName,
  media_type: 'application/pdf',
  size: 9_876_543,
  uploaded_by: staff,
  uploaded_at: '2026-11-17T15:00:00Z',
  current: true,
})

const demo = detail(
  [field('reason', 'Motivo', 'long_text'), field('comment', 'Comentario', 'text')],
  [
    text('reason', 'Pido la equivalencia de una materia cursada en otra facultad.'),
    text('comment', 'Sin comentarios.'),
  ],
  [document('ruling', 'Dictamen')],
  [started],
)

// The longest values the server accepts (200 characters in a text, 5000 in a
// long text, 200 in a file name), a long list of subjects and long names.
const subjects = Array.from(
  { length: 24 },
  (_, i) => `Análisis Matemático ${i + 1} (Plan 2011, código 61.${String(i + 3).padStart(2, '0')})`,
)
const longFileName =
  'Dictamen_Comision_Curricular_Equivalencias_2026_Expediente_EX-2026-00012345-UBA-DME_FI_Resolucion_CD_N_1234-2026_version_final_firmada_por_todos_los_integrantes_de_la_comision_escaneada.pdf'
const worst = detail(
  [
    field(
      'reason',
      'Motivo por el cual solicita el reconocimiento de las asignaturas cursadas en otra unidad académica',
      'long_text',
    ),
    field('url', 'Enlace al programa de la asignatura', 'text'),
    field('subjects', 'Asignaturas', 'subject_list'),
    field('w', 'W', 'text'),
  ],
  [
    text(
      'reason',
      Array.from(
        { length: 12 },
        (_, i) =>
          `Párrafo ${i + 1}: cursé y aprobé la asignatura en la Facultad de Ciencias Exactas e Ingeniería de otra universidad nacional, con una carga horaria de 128 horas y un programa equivalente al vigente en esta facultad, según consta en la documentación adjunta.`,
      ).join('\n\n'),
    ),
    text(
      'url',
      'https://campus.unidad-academica-de-ejemplo.edu.ar/programas/2019/analisis-matematico-ii-programa-completo-con-bibliografia-y-cronograma-de-clases-del-primer-cuatrimestre-version-final.pdf',
    ),
    { ...text('subjects', ''), text: undefined, items: subjects },
    text('w', 'Ñ'),
  ],
  [
    document('ruling', 'Dictamen de la Comisión Curricular de Equivalencias y Reconocimiento de Créditos Académicos'),
    document('record', 'Acta'),
    document('annex', 'Anexo con la tabla de correspondencias entre asignaturas', false),
  ],
  [started, attached('D-1', 'ruling', longFileName), attached('D-2', 'unknown', 'acta.pdf')],
  [file('D-1', 'ruling', longFileName)],
)

const empty = detail([field('reason', 'Motivo', 'long_text')], [], [], [])

const one = detail(
  [field('reason', 'Motivo', 'text')],
  [text('reason', 'A')],
  [document('ruling', 'Dictamen')],
  [started],
)

const review: Step = {
  id: 'review',
  name: 'Revisión de la solicitud',
  action: 'decide',
  role: 'staff',
  area: { id: 'office', name: 'Dirección de Alumnos' },
  fields: [],
  documents: [],
  outcomes: [
    { id: 'accept', name: 'Aceptar', next: 'communicate', send_back: false },
    { id: 'send_back', name: 'Devolver al titular', next: 'request', send_back: true },
  ],
  values: [],
  files: [],
}

const sendBack = (observation: string): Event => ({
  step: 'review',
  by: staff,
  at: '2026-11-17T15:00:00Z',
  observation,
  outcome: 'send_back',
  declared: false,
  send_back: true,
  next: 'request',
  returned: false,
})

// The staff deciding on a step that can send the procedure back.
const deciding: Detail = {
  ...demo,
  steps: [demo.steps[0]!, review, demo.steps[1]!],
  open_step: 'review',
}

// The applicant, who has the procedure again after a send back: kind reason,
// and the longest one the server accepts (5000 characters).
const sentBack: Detail = {
  ...deciding,
  state: 'with_applicant',
  open_step: 'request',
  can_act: false,
  history: [started, sendBack('Falta la firma en la solicitud.')],
}
const sentBackWorst: Detail = {
  ...sentBack,
  history: [
    started,
    sendBack(
      Array.from(
        { length: 12 },
        (_, i) =>
          `Párrafo ${i + 1}: la solicitud no incluye la firma del titular ni el sello de la unidad académica de origen, y el programa de la asignatura no coincide con el que figura en la constancia adjunta.`,
      ).join('\n\n'),
    ),
  ],
}

export const fixtures = { demo, worst, empty, one, deciding, sentBack, sentBackWorst }

export type FixtureName = keyof typeof fixtures
