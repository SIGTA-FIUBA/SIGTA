import type { components } from '../api/schema'

type Area = components['schemas']['Area']
type AreaMembership = components['schemas']['AreaMembership']
type Administrator = components['schemas']['Administrator']
type CatalogEntry = components['schemas']['CatalogEntry']

// What the administration page reads, for AdministrationPreview. memberships
// goes by area. The names are made up: no fixture holds real people.
type Fixture = {
  areas: Area[]
  memberships: Record<string, AreaMembership[]>
  administrators: Administrator[]
  procedureTypes: CatalogEntry[]
}

const secretariat: Area = {
  id: 'secretariat',
  name: 'Secretaría',
  system: true,
  permissions: [
    { id: 'edit_procedure_types', name: 'Editar tipos de trámite', system: true },
    { id: 'publish_procedure_types', name: 'Publicar tipos de trámite', system: true },
  ],
  roles: [],
}

const office: Area = {
  id: 'office',
  name: 'Dirección de Alumnos',
  system: false,
  permissions: [
    { id: 'review', name: 'Revisar', system: false },
    { id: 'register', name: 'Registrar expediente', system: false },
  ],
  roles: [{ id: 'clerk', name: 'Administrativo', permissions: ['review', 'register'] }],
}

const procedureType = (id: number, name: string): CatalogEntry => ({
  id,
  version: 1,
  name,
  description: '',
  deadlines: [],
  admissibility: [],
  fields: [],
  documents: [],
  submission: { fields: [], documents: [] },
})

const firstAdministrator: Administrator = {
  id: 'a1',
  person: { id: 'u1', name: 'Persona Uno', email: 'persona.uno@fi.uba.ar' },
  from: '2026-10-01T12:00:00Z',
}

const demo: Fixture = {
  areas: [secretariat, office],
  memberships: {
    secretariat: [],
    office: [
      {
        id: 'm1',
        person: { id: 'u2', name: 'Persona Dos', email: 'persona.dos@fi.uba.ar' },
        roles: ['clerk'],
        permissions: [],
        scope: [],
        from: '2026-11-01T12:00:00Z',
        granted_by: { id: 'u1', name: 'Persona Uno' },
      },
    ],
  },
  administrators: [
    firstAdministrator,
    {
      id: 'a2',
      person: { id: 'u2', name: 'Persona Dos', email: 'persona.dos@fi.uba.ar' },
      from: '2026-11-02T12:00:00Z',
      granted_by: { id: 'u1', name: 'Persona Uno' },
    },
  ],
  procedureTypes: [procedureType(1, 'Equivalencia')],
}

const longArea: Area = {
  id: 'comision_curricular_de_ingenieria_en_informatica_y_licenciatura_en_analisis_de_sistemas',
  name: 'Comisión Curricular Permanente de Ingeniería en Informática y Licenciatura en Análisis de Sistemas',
  system: false,
  permissions: Array.from({ length: 14 }, (_, i) => ({
    id: `dictaminar_equivalencias_de_materias_de_otras_universidades_${i + 1}`,
    name: `Dictaminar equivalencias de materias aprobadas en otras universidades nacionales ${i + 1}`,
    system: false,
  })),
  roles: [
    {
      id: 'integrante_titular_de_la_comision',
      name: 'Integrante titular de la comisión con voto en los dictámenes',
      permissions: Array.from(
        { length: 14 },
        (_, i) => `dictaminar_equivalencias_de_materias_de_otras_universidades_${i + 1}`,
      ),
    },
    { id: 'jo', name: 'Jo', permissions: ['dictaminar_equivalencias_de_materias_de_otras_universidades_1'] },
  ],
}

const worst: Fixture = {
  areas: [secretariat, longArea, office],
  memberships: {
    secretariat: [],
    [longArea.id]: [
      {
        id: 'w1',
        person: {
          id: 'u3',
          name: 'Aleksandra Wiśniewska-Kowalczyk de los Santos Montgomery',
          email: 'bartholomew.fitzgerald.wisniewska-kowalczyk.montgomery@fi.uba.ar',
        },
        roles: ['integrante_titular_de_la_comision', 'jo'],
        permissions: longArea.permissions.slice(0, 6).map((permission) => permission.id),
        scope: [1, 2, 3, 4, 5, 6, 99],
        from: '2026-11-01T12:00:00Z',
        until: '2027-04-01T02:59:59Z',
        granted_by: { id: 'u9', name: '' },
      },
      {
        id: 'w2',
        person: { id: 'u4', name: 'Jo', email: 'jo@fi.uba.ar' },
        roles: [],
        permissions: ['retired_permission'],
        scope: [],
        from: '2027-01-15T12:00:00Z',
        granted_by: { id: 'u1', name: 'Persona Uno' },
      },
    ],
    office: [],
  },
  administrators: [
    firstAdministrator,
    {
      id: 'a3',
      person: {
        id: 'u3',
        name: 'Aleksandra Wiśniewska-Kowalczyk de los Santos Montgomery',
        email: 'bartholomew.fitzgerald.wisniewska-kowalczyk.montgomery@fi.uba.ar',
      },
      from: '2026-11-02T12:00:00Z',
      granted_by: { id: 'u9', name: '' },
    },
  ],
  procedureTypes: [
    procedureType(1, 'Equivalencia'),
    procedureType(2, 'Reconocimiento de créditos por actividades de extensión universitaria y voluntariado'),
    procedureType(3, 'Créditos'),
    procedureType(4, 'Cambio de carrera'),
    procedureType(5, 'Simultaneidad'),
    procedureType(6, 'Readmisión'),
  ],
}

const empty: Fixture = {
  areas: [{ ...secretariat }],
  memberships: { secretariat: [] },
  administrators: [firstAdministrator],
  procedureTypes: [],
}

export const fixtures = { demo, worst, empty }

export type FixtureName = keyof typeof fixtures
