import { AppShell, Button, Paper } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { createMemoryRouter, Link, RouterProvider } from 'react-router'
import type { components } from '../api/schema'
import { AppHeader } from './AppHeader'
import { fixtures } from '../dev/administrationFixtures'
import { fixtures as procedureFixtures } from '../dev/procedureFixtures'
import { AdministrationPage } from '../pages/AdministrationPage'
import { HomePage } from '../pages/HomePage'
import { InboxPage } from '../pages/InboxPage'
import { LoginPage } from '../pages/LoginPage'
import { ProcedurePage } from '../pages/ProcedurePage'
import { tokens } from '../theme/tokens'
import { AppLayout } from './AppLayout'
import { UserMenu } from './UserMenu'

type User = components['schemas']['User']
type ProcedureList = components['schemas']['ProcedureList']
type MyProcedure = components['schemas']['MyProcedure']
type CatalogEntry = components['schemas']['CatalogEntry']

// The frame of the pages runs inside a router, as in the app. The session,
// the inbox and the procedures of the person are invented, so the stories
// make no requests.
const staff: User = {
  id: 'u1',
  email: 'operativo@ejemplo.edu.ar',
  name: 'Laura Méndez',
  staff: true,
  administrator: false,
}
const applicant: User = {
  id: 'u2',
  email: 'titular@ejemplo.edu.ar',
  name: 'Martín Sosa',
  staff: false,
  administrator: false,
}

// The inbox: procedures of several types, areas and states, some of them
// waiting on the account. Every one opens the same example detail, with the
// name of its type.
const types = [
  { id: 1, name: 'Equivalencia entre asignaturas' },
  { id: 2, name: 'Reconocimiento de créditos' },
  { id: 3, name: 'Pase de carrera' },
  { id: 4, name: 'Pase desde una universidad extranjera' },
]
const areas = [
  { id: 'secretaria', name: 'Secretaría Académica' },
  { id: 'alumnos', name: 'Dirección de Alumnos' },
  { id: 'computacion', name: 'Departamento de Computación' },
]
const rows = [
  ['Martín Sosa', 0, 'Revisión de la Secretaría', 0, 'in_office', true, 12, 1],
  ['Valeria Ríos', 1, 'Dictamen del departamento', 2, 'referred', false, 20, 3],
  ['Tomás Aguirre', 2, 'Registrar expediente', 1, 'in_office', true, 4, 0],
  ['Julieta Paz', 0, 'Documentación', 0, 'with_applicant', false, 9, 2],
  ['Nicolás Ferreyra', 3, 'Revisión de la Secretaría', 0, 'in_office', true, 31, 5],
  ['Camila Ortiz', 1, 'Comunicación del dictamen', 1, 'in_office', true, 15, 4],
  ['Bruno Medina', 2, 'Dictamen del departamento', 2, 'referred', false, 7, 6],
] as const

const day = 86_400_000
const inbox: ProcedureList = {
  items: rows.map(([name, type, step, area, state, actionable, age, moved], index) => ({
    id: `p${index + 1}`,
    procedure_type: types[type]!.id,
    procedure_type_name: types[type]!.name,
    applicant: { id: `a${index + 1}`, name },
    step_name: step,
    area: state === 'with_applicant' ? undefined : areas[area],
    state,
    actionable,
    started_at: new Date(Date.now() - age * day).toISOString(),
    last_moved_at: new Date(Date.now() - moved * day).toISOString(),
    age_days: age,
  })),
  procedure_types: types.map((type) => ({ id: String(type.id), name: type.name })),
  areas,
}

// The procedures the person started: one waiting for a correction and one
// closed.
const mine: MyProcedure[] = [
  {
    id: 'm1',
    procedure_type: 2,
    procedure_type_name: 'Reconocimiento de créditos',
    step_name: 'Documentación',
    state: 'with_applicant',
    correction_pending: true,
    started_at: '2026-11-02T12:00:00Z',
    last_moved_at: '2026-11-12T15:00:00Z',
  },
  {
    id: 'm2',
    procedure_type: 3,
    procedure_type_name: 'Pase de carrera',
    state: 'closed',
    correction_pending: false,
    started_at: '2026-10-01T12:00:00Z',
    last_moved_at: '2026-11-05T15:00:00Z',
  },
]

// seedAdministration puts the demo areas, memberships and administrators of
// the development preview where the administration reads them.
function seedAdministration(client: QueryClient) {
  const data = fixtures.demo
  client.setQueryData(['areas'], data.areas)
  for (const [area, memberships] of Object.entries(data.memberships)) {
    client.setQueryData(['areas', area, 'memberships'], memberships)
  }
  client.setQueryData(['administrators'], data.administrators)
  client.setQueryData(['procedure-types'], data.procedureTypes)
}

function Page({ user, path = '/', children }: { user: User | null; path?: string; children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { staleTime: Infinity, retry: false } } })
  client.setQueryData(['procedures'], inbox)
  for (const item of inbox.items) {
    client.setQueryData(['procedure', item.id], {
      ...procedureFixtures.demo,
      id: item.id,
      name: item.procedure_type_name,
    })
  }
  client.setQueryData(['me', 'procedures'], mine)
  seedAdministration(client)
  const router = createMemoryRouter(
    [
      {
        path: '/*',
        loader: () => user,
        element: <AppLayout />,
        children: [
          { path: 'procedures/:id', element: <ProcedurePage /> },
          { path: '*', element: children },
        ],
      },
    ],
    { initialEntries: [path] },
  )
  return (
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
}

// The landing page counts the catalog; three invented types are enough.
const catalog: CatalogEntry[] = ['Equivalencia entre asignaturas', 'Reconocimiento de créditos', 'Pase de carrera'].map(
  (name, index) => ({
    id: index + 1,
    version: 1,
    name,
    description: '',
    deadlines: [],
    admissibility: [],
    fields: [],
    documents: [],
    submission: { fields: [], documents: [] },
  }),
)

function Landing() {
  const client = new QueryClient({ defaultOptions: { queries: { staleTime: Infinity, retry: false } } })
  client.setQueryData(['procedure-types'], catalog)
  const router = createMemoryRouter([{ path: '*', element: <LoginPage /> }], { initialEntries: ['/login'] })
  return (
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
}

const meta: Meta = {
  title: 'SIGTA/Marco',
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component:
          'El marco de todas las páginas (`src/layout/`): con sesión, un menú lateral blanco de 240 de arriba abajo, con un borde a la derecha, con el logo de SIGTA, las opciones y, al pie, Acerca de SIGTA y la cuenta. En una tableta el menú queda en íconos con tooltip; en el celular se pliega detrás de un botón, en una barra superior blanca. Sin sesión, solo la barra superior con el ingreso.',
      },
    },
  },
}

export default meta

export const AppLayoutStaff: StoryObj = {
  name: 'AppLayout, personal',
  render: () => <Page user={staff}>{null}</Page>,
  parameters: {
    docs: {
      description: {
        story: 'Con sesión del personal: menú con la bandeja y la cantidad de trámites con acción requerida.',
      },
    },
  },
}

export const AppLayoutApplicant: StoryObj = {
  name: 'AppLayout, titular',
  render: () => <Page user={applicant}>{null}</Page>,
}

export const AppLayoutVisitor: StoryObj = {
  name: 'AppLayout, sin sesión',
  render: () => <Page user={null}>{null}</Page>,
  parameters: {
    docs: { description: { story: 'Sin sesión (el catálogo público): sin menú, con el ingreso en el encabezado.' } },
  },
}

export const LandingStory: StoryObj = {
  name: 'Portada',
  render: () => <Landing />,
  parameters: {
    docs: {
      description: {
        story:
          'La portada pública (`/login`), antes del ingreso: el titular con el botón de Google y la nota de la cuenta de la UBA, un collage de piezas de la plataforma (se oculta en el celular), cómo funciona, cuántos trámites hay en el catálogo y los criterios. Es la única pantalla con píldoras y títulos que crecen con el ancho.',
      },
    },
  },
}

export const HomeStaff: StoryObj = {
  name: 'Inicio, personal',
  render: () => (
    <Page user={staff}>
      <HomePage />
    </Page>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'El Inicio del personal: el saludo con la fecha, el resumen en dos cifras (cada una lleva a la bandeja filtrada), hasta cinco trámites que requieren acción y la actividad reciente, un renglón por movimiento.',
      },
    },
  },
}

export const HomeApplicant: StoryObj = {
  name: 'Inicio, titular',
  render: () => (
    <Page user={applicant}>
      <HomePage />
    </Page>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'El Inicio del titular: el saludo, sus trámites en curso y cerrados, y los pasos de un trámite. El catálogo está en el menú, así que ningún botón lo repite.',
      },
    },
  },
}

export const InboxStaff: StoryObj = {
  name: 'Bandeja, personal',
  render: () => (
    <Page user={staff} path="/inbox">
      <InboxPage />
    </Page>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'La bandeja: las cifras por estado, que filtran la lista (la marca de la seleccionada se desliza a la elegida), y los trámites, que entran uno tras otro y viajan a su lugar cuando cambia el orden. Al abrir uno, el tipo de trámite pasa a ser el título de su página.',
      },
    },
  },
}

export const AdministrationStory: StoryObj = {
  name: 'Administración',
  render: () => (
    <Page user={{ ...staff, administrator: true }} path="/administration">
      <AdministrationPage />
    </Page>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'La administración, con los datos de ejemplo de la vista de desarrollo: áreas con sus permisos, roles y membresías, y las personas administradoras en otra pestaña. El menú suma Administración solo para quien administra.',
      },
    },
  },
}

export const AppHeaderStory: StoryObj = {
  name: 'AppHeader',
  // AppHeader fills the height of its bar, so it goes inside an AppShell header
  // with the height AppLayout gives it.
  render: () => (
    <RouterProvider
      router={createMemoryRouter([
        {
          path: '/',
          element: (
            <AppShell header={{ height: tokens.layout.topBar }} mode="static">
              <AppShell.Header>
                <AppHeader
                  end={
                    <Button component={Link} to="/" variant="subtle">
                      Una acción
                    </Button>
                  }
                />
              </AppShell.Header>
            </AppShell>
          ),
        },
      ])}
    />
  ),
}

export const UserMenuStory: StoryObj = {
  name: 'UserMenu',
  render: () => (
    <Paper p="sm" w={tokens.layout.nav} withBorder>
      <UserMenu user={staff} onLogout={() => {}} />
    </Paper>
  ),
  parameters: { layout: 'centered' },
}
