import type { components } from '../api/schema'

type ProcedureTypeDraft = components['schemas']['ProcedureTypeDraft']
type EditableProcedureTypes = components['schemas']['EditableProcedureTypes']

const laura = { id: '5RnF8i6-BEyFCTX4fs2BN-WCnyQNj5K89Y_xKGfvQvU', name: 'Laura Méndez' }
const martin = { id: '3Las2QqynzRcykV5BlT4Pm4CDB_E3jrmzs4YVUfsc8o', name: 'Martín Ruiz' }

// What the editor of procedure types reads, for ProcedureTypeDraftPreview.
// The names are made up: no fixture holds real people. The identifiers are
// opaque, as those of Google accounts, so that a screen that showed them
// instead of the names would show here.
type Fixture = { draft: ProcedureTypeDraft; editable: EditableProcedureTypes }

const editable: EditableProcedureTypes = {
  can_create: true,
  procedure_types: [],
  areas: [
    { id: 'desk', name: 'Mesa de entradas', permissions: [{ id: 'verify', name: 'Verificar' }] },
    { id: 'commission', name: 'Comisión Curricular', permissions: [{ id: 'evaluate', name: 'Dictaminar' }] },
    { id: 'secretariat', name: 'Secretaría', permissions: [{ id: 'resolve', name: 'Resolver' }] },
  ],
}

const definition: ProcedureTypeDraft['definition'] = {
  name: 'Equivalencia entre asignaturas',
  description: '',
  deadlines: ['Se presenta hasta 30 días después de la inscripción a la carrera.'],
  admissibility: ['Las asignaturas están aprobadas en otra carrera.'],
  fields: [
    { id: 'subjects', name: 'Materias a reconocer', type: 'subject_list' },
    { id: 'plan', name: 'Plan de estudios', type: 'text' },
    { id: 'file_number', name: 'Número de expediente', type: 'text' },
  ],
  documents: [
    { id: 'transcript', name: 'Certificado analítico', accepted_types: ['application/pdf'], max_size: 5 * 1024 * 1024 },
    { id: 'ruling', name: 'Dictamen', accepted_types: ['application/pdf'], max_size: 5 * 1024 * 1024 },
    {
      id: 'resolution',
      name: 'Resolución',
      accepted_types: ['application/pdf'],
      max_size: 5 * 1024 * 1024,
      resolution: true,
    },
  ],
  steps: [
    {
      id: 'request',
      name: 'Solicitud',
      role: 'applicant',
      action: 'enter_data',
      fields: ['subjects'],
      next: 'documents',
    },
    {
      id: 'documents',
      name: 'Documentación',
      role: 'applicant',
      action: 'upload_document',
      documents: ['transcript'],
      next: 'verification',
    },
    {
      id: 'verification',
      name: 'Verificación',
      role: 'staff',
      area: 'desk',
      permissions: ['verify'],
      action: 'confirm_external_fact',
      fields: ['file_number'],
      next: 'ruling',
    },
    {
      id: 'ruling',
      name: 'Dictamen de la comisión',
      role: 'staff',
      permissions: [],
      action: 'decide',
      documents: ['ruling'],
      outcomes: [
        { id: 'go_on', name: 'Continuar', next: 'resolution', send_back: false },
        { id: 'ask', name: 'Pedir información', next: 'documents', send_back: true },
      ],
    },
    {
      id: 'resolution',
      name: 'Resolución',
      role: 'staff',
      area: 'secretariat',
      permissions: ['resolve'],
      action: 'upload_document',
      documents: ['resolution'],
    },
  ],
}

const draft: ProcedureTypeDraft = {
  procedure_type_id: 1,
  base: 3,
  revision: 7,
  status: 'editing',
  definition,
  edited_by: laura,
  edited_at: '2026-11-21T16:30:00Z',
  editors: [laura],
  can_edit: true,
  can_review: false,
  problems: [
    { code: 'no_description', path: [{ kind: 'definition' }], text: 'definition has no description' },
    { code: 'unused_field', path: [{ kind: 'field', id: 'plan' }], text: 'field "plan" is not used by any step' },
    {
      code: 'no_permissions',
      path: [{ kind: 'step', id: 'ruling' }],
      text: 'staff steps ask for at least one permission',
    },
  ],
}

const long = 'Reconocimiento de asignaturas aprobadas en universidades extranjeras con convenio vigente'

export const fixtures = {
  demo: { draft, editable },
  proposed: {
    draft: {
      ...draft,
      status: 'proposed',
      problems: [],
      proposed_by: laura,
      proposed_at: '2026-11-21T16:40:00Z',
      can_edit: false,
      can_review: true,
    },
    editable,
  },
  worst: {
    draft: {
      ...draft,
      definition: {
        ...definition,
        name: long,
        steps: [
          ...definition.steps.slice(0, 3),
          {
            ...definition.steps[3],
            name: `Dictamen de la ${long}`,
            outcomes: [
              { id: 'go_on', name: 'Continuar con la resolución del trámite', next: 'resolution', send_back: false },
              { id: 'ask', name: 'Pedir información adicional al titular', next: 'request', send_back: true },
              { id: 'back', name: 'Volver a verificar', next: 'verification', send_back: true },
              { id: 'archive', name: 'Archivar', next: 'resolution', send_back: false },
            ],
          },
          definition.steps[4],
        ],
      },
      rejection: {
        by: martin,
        at: '2026-11-21T17:05:00Z',
        reason: 'Falta el plazo de presentación y el área del dictamen. '.repeat(4),
      },
      problems: [...draft.problems, ...draft.problems, ...draft.problems],
    },
    editable,
  },
} satisfies Record<string, Fixture>

export type FixtureName = keyof typeof fixtures
