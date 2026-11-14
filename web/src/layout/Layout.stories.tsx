import { AppShell, Button } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { createMemoryRouter, Link, RouterProvider } from 'react-router'
import type { components } from '../api/schema'
import { AppHeader } from './AppHeader'
import { AppLayout } from './AppLayout'
import { UserMenu } from './UserMenu'

type User = components['schemas']['User']
type ProcedureList = components['schemas']['ProcedureList']

// The frame of the pages runs inside a router, as in the app. The session
// and the inbox count are invented, so the stories make no requests.
const staff: User = { id: 'u1', email: 'operativo@ejemplo.edu.ar', name: 'Laura Méndez', role: 'staff_operative' }
const applicant: User = { id: 'u2', email: 'titular@ejemplo.edu.ar', name: 'Martín Sosa', role: 'applicant' }

function procedures(actionable: number): ProcedureList {
  const item = {
    procedure_type: 1,
    procedure_type_name: 'Equivalencia entre asignaturas',
    applicant_name: 'Martín Sosa',
    step_name: 'Revisión de la Secretaría',
    area: 'secretaria',
    area_name: 'Secretaría Académica',
    state: 'in_office' as const,
    started_at: '2026-11-02T10:00:00Z',
    last_moved_at: '2026-11-10T10:00:00Z',
    age_days: 12,
  }
  return {
    items: Array.from({ length: actionable }, (_, index) => ({ ...item, id: `p${index}`, actionable: true })),
    procedure_types: [],
    areas: [],
  }
}

function Page({ user, children }: { user: User | null; children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { staleTime: Infinity, retry: false } } })
  client.setQueryData(['procedures'], procedures(3))
  const router = createMemoryRouter(
    [{ path: '/', loader: () => user, element: <AppLayout />, children: [{ index: true, element: children }] }],
    { initialEntries: ['/'] },
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
          'El marco de todas las páginas (`src/layout/`): el encabezado con el logo y las acciones, el menú lateral celeste con sesión y el menú de la cuenta.',
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
            <AppShell header={{ height: 64 }} mode="static">
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
  render: () => <UserMenu user={staff} onLogout={() => {}} />,
  parameters: { layout: 'centered' },
}
