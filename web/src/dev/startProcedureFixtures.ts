import type { components } from '../api/schema'

// Catalogs for the preview of the form that starts a procedure: kind data, the
// realistic worst case, and the edge counts. Synthetic data only (ADR-005).

type CatalogEntry = components['schemas']['CatalogEntry']
type FieldRequirement = components['schemas']['FieldRequirement']
type DocumentRequirement = components['schemas']['DocumentRequirement']

const megabyte = 1024 * 1024

function entry(name: string, fields: FieldRequirement[], documents: DocumentRequirement[]): CatalogEntry {
  return {
    id: 1,
    version: 1,
    name,
    description: '',
    deadlines: [],
    admissibility: [],
    fields,
    documents,
    submission: { fields, documents },
  }
}

const field = (
  id: string,
  name: string,
  type: FieldRequirement['type'],
  required = true,
  options: FieldRequirement['options'] = [],
): FieldRequirement => ({ id, name, type, required, options })

const document = (
  id: string,
  name: string,
  acceptedTypes: string[],
  maxSize: number,
  required = true,
): DocumentRequirement => ({ id, name, required, accepted_types: acceptedTypes, max_size: maxSize, resolution: false })

const demo = entry(
  'Equivalencia entre asignaturas',
  [
    field('reason', 'Motivo', 'long_text'),
    field('modality', 'Modalidad', 'option', true, [
      { id: 'full', name: 'Total' },
      { id: 'partial', name: 'Parcial' },
    ]),
  ],
  [document('transcript', 'Historia académica', ['application/pdf'], 10 * megabyte)],
)

// Degrees as a configuration would list them: many options, long names.
const degrees = [
  'Ingeniería Civil',
  'Ingeniería Electricista',
  'Ingeniería Electrónica',
  'Ingeniería en Agrimensura',
  'Ingeniería en Alimentos',
  'Ingeniería en Energía Eléctrica',
  'Ingeniería en Informática',
  'Ingeniería en Petróleo',
  'Ingeniería Industrial',
  'Ingeniería Mecánica',
  'Ingeniería Naval y Mecánica',
  'Ingeniería Química',
  'Licenciatura en Análisis de Sistemas',
  'Licenciatura en Ciencias de la Computación (articulación con la Facultad de Ciencias Exactas y Naturales)',
].map((name, i) => ({ id: `degree_${i + 1}`, name }))

const worst = entry(
  'Reconocimiento de créditos por actividades de extensión, investigación, transferencia y prácticas profesionales supervisadas realizadas fuera de la Facultad',
  [
    field(
      'reason',
      'Motivo de la solicitud, con el detalle de las asignaturas aprobadas en la institución de origen y su correspondencia con el plan vigente',
      'long_text',
    ),
    field('origin_degree', 'Carrera de destino', 'option', true, degrees),
    field('credits', 'Créditos', 'number'),
    field('since', 'Fecha de aprobación de la última asignatura', 'date'),
    field('subjects', 'Asignaturas aprobadas', 'subject_list'),
    field('institution', 'Institución', 'text', false),
    field('contact', 'Correo alternativo', 'text', false),
  ],
  [
    document(
      'transcript',
      'Certificado analítico legalizado',
      ['application/pdf', 'image/png', 'image/jpeg'],
      20 * megabyte,
    ),
    document(
      'programs',
      'Programas analíticos de cada asignatura, sellados y firmados por la institución de origen',
      ['application/pdf'],
      20 * megabyte,
    ),
    document('passport', 'Pasaporte', ['application/pdf', 'image/png'], 2 * megabyte, false),
    document('legacy', 'Constancia de un formato anterior', [], 0, false),
  ],
)

const empty = entry('Baja de una inscripción', [], [])

const one = entry(
  'Pedido de constancia',
  [field('modality', 'Modalidad', 'option', true, [{ id: 'only', name: 'Única' }])],
  [document('note', 'Nota', ['application/pdf'], 1024)],
)

export const fixtures = { demo, worst, empty, one } as const
export type FixtureName = keyof typeof fixtures
