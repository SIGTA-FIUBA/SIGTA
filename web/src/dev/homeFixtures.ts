import type { components } from '../api/schema'

// homeFixtures are the procedures of the person on the home page, closed ones
// with their resolutions: kind data, the realistic worst case, and the edge
// counts. Synthetic data only (ADR-005).

type MyProcedure = components['schemas']['MyProcedure']
type ResolutionFile = components['schemas']['ResolutionFile']

function resolution(
  id: string,
  documentName: string,
  fileName: string,
  mediaType: string,
  size: number,
): ResolutionFile {
  return { id, document: id, document_name: documentName, file_name: fileName, media_type: mediaType, size }
}

function closed(id: string, name: string, lastMovedAt: string, resolutions: ResolutionFile[]): MyProcedure {
  return {
    id,
    procedure_type: 1,
    procedure_type_name: name,
    state: 'closed',
    correction_pending: false,
    started_at: '2026-09-01T12:00:00Z',
    last_moved_at: lastMovedAt,
    resolutions,
  }
}

const inOffice: MyProcedure = {
  id: 'P-1',
  procedure_type: 1,
  procedure_type_name: 'Equivalencia entre asignaturas',
  step_name: 'Revisión de la solicitud',
  state: 'in_office',
  correction_pending: false,
  started_at: '2026-11-02T12:00:00Z',
  last_moved_at: '2026-11-10T15:00:00Z',
  resolutions: [],
}

const demo: MyProcedure[] = [
  inOffice,
  closed('P-2', 'Pase de carrera', '2026-11-05T15:00:00Z', [
    resolution('dean_resolution', 'Resolución de Decano', 'resolucion.pdf', 'application/pdf', 184_320),
  ]),
]

// The file name is as long as the server keeps (maxFileName, 200), without
// spaces, as a scanner or an expedient system names it.
const worst: MyProcedure[] = [
  {
    ...inOffice,
    procedure_type_name:
      'Reconocimiento de créditos por actividades académicas, de investigación y de extensión universitaria',
    correction_pending: true,
    state: 'with_applicant',
    step_name: 'Documentación complementaria solicitada por la Comisión Curricular',
  },
  closed(
    'P-2',
    'Reconocimiento de créditos por actividades académicas, de investigación y de extensión universitaria',
    '2026-11-20T15:00:00Z',
    [
      resolution(
        'council_resolution',
        'Resolución del Consejo Directivo con el dictamen de la Comisión Curricular de Ingeniería en Informática',
        'RESCD_2026_01842_UBA_DCT_FI_Reconocimiento_de_creditos_actividades_de_investigacion_y_extension_EX_2026_00012345_UBA_DME_SA_FI_firmada_digitalmente_con_anexo_de_asignaturas_reconocidas_version_final_3.pdf',
        'application/pdf',
        26_214_400,
      ),
      resolution(
        'annex',
        'Anexo',
        'anexo.docx',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        0,
      ),
      resolution('notice', 'Notificación al titular', 'IMG_20261120_183022_HDR.jpg', 'image/jpeg', 4_718_592),
    ],
  ),
  closed('P-3', 'Pase', '2026-11-19T15:00:00Z', []),
  closed(
    'P-4',
    'Intercambio académico con universidades del exterior (programa de doble titulación)',
    '2026-11-18T15:00:00Z',
    [resolution('dean_resolution', 'R', 'r.pdf', 'application/pdf', 1)],
  ),
]

const one: MyProcedure[] = [
  closed('P-2', 'Pase de carrera', '2026-11-05T15:00:00Z', [
    resolution('dean_resolution', 'Resolución de Decano', 'resolucion.pdf', 'application/pdf', 1_024),
  ]),
]

const many: MyProcedure[] = Array.from({ length: 40 }, (_, i) =>
  closed(`P-${i + 1}`, i % 2 ? 'Pase de carrera' : 'Equivalencia entre asignaturas', '2026-11-05T15:00:00Z', [
    resolution('dean_resolution', 'Resolución de Decano', `resolucion-${i + 1}.pdf`, 'application/pdf', 184_320),
  ]),
)

export const fixtures = { demo, worst, empty: [] as MyProcedure[], one, many }

export type FixtureName = keyof typeof fixtures
