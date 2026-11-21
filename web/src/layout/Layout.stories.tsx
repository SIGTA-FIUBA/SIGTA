import { AppShell, Button, Paper } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { createMemoryRouter, Link, RouterProvider } from 'react-router'
import type { components } from '../api/schema'
import { AppHeader } from './AppHeader'
import { fixtures } from '../dev/administrationFixtures'
import { AdministrationPage } from '../pages/AdministrationPage'
import { HomePage } from '../pages/HomePage'
import { tokens } from '../theme/tokens'
import { AppLayout } from './AppLayout'
import { UserMenu } from './UserMenu'

type User = components['schemas']['User']
type ProcedureList = components['schemas']['ProcedureList']
type MyProcedure = components['schemas']['MyProcedure']

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

const states = ['in_office', 'referred', 'with_applicant'] as const

function procedures(actionable: number): ProcedureList {
  const item = {
    procedure_type: 1,
    procedure_type_name: 'Equivalencia entre asignaturas',
    applicant: { id: 'u2', name: 'Martín Sosa' },
    step_name: 'Revisión de la Secretaría',
    area: { id: 'secretaria', name: 'Secretaría Académica' },
    state: 'in_office' as const,
    started_at: '2026-11-02T10:00:00Z',
    age_days: 12,
  }
  return {
    items: Array.from({ length: actionable }, (_, index) => ({
      ...item,
      id: `p${index}`,
      state: states[index % states.length]!,
      last_moved_at: `2026-11-${String(10 - index).padStart(2, '0')}T10:00:00Z`,
      actionable: true,
    })),
    procedure_types: [],
    areas: [],
  }
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
  client.setQueryData(['procedures'], procedures(3))
  client.setQueryData(['me', 'procedures'], mine)
  seedAdministration(client)
  const router = createMemoryRouter(
    [{ path: '/*', loader: () => user, element: <AppLayout />, children: [{ path: '*', element: children }] }],
    { initialEntries: [path] },
  )
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
