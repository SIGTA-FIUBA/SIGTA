import { AppShell, Button, Paper } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { createMemoryRouter, Link, RouterProvider } from 'react-router'
import type { components } from '../api/schema'
import { AppHeader } from './AppHeader'
import { HomePage } from '../pages/HomePage'
import { tokens } from '../theme/tokens'
import { AppLayout } from './AppLayout'
import { UserMenu } from './UserMenu'

type User = components['schemas']['User']
type ProcedureList = components['schemas']['ProcedureList']

// The frame of the pages runs inside a router, as in the app. The session
// and the inbox count are invented, so the stories make no requests.
const staff: User = { id: 'u1', email: 'operativo@ejemplo.edu.ar', name: 'Laura Méndez', role: 'staff_operative' }
const applicant: User = { id: 'u2', email: 'titular@ejemplo.edu.ar', name: 'Martín Sosa', role: 'applicant' }

const states = ['in_office', 'referred', 'with_applicant'] as const

function procedures(actionable: number): ProcedureList {
  const item = {
    procedure_type: 1,
    procedure_type_name: 'Equivalencia entre asignaturas',
    applicant_name: 'Martín Sosa',
    step_name: 'Revisión de la Secretaría',
    area: 'secretaria',
    area_name: 'Secretaría Académica',
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
          'El marco de todas las páginas (`src/layout/`): con sesión, un menú lateral azul marino de 240 de arriba abajo, con el logo de SIGTA, las opciones y, al pie, Acerca de SIGTA y la cuenta. En una tableta el menú queda en íconos con tooltip; en el celular se pliega detrás de un botón, en una barra superior blanca. Sin sesión, solo la barra superior con el ingreso.',
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
          'El Inicio del personal: la bandeja como acción principal y los trámites que requieren acción, por última novedad.',
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
          'El Inicio del titular: iniciar un trámite como acción principal y los pasos de un trámite. Falta "Mis trámites": todavía no hay forma de listar los trámites de la persona.',
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
    <Paper p="sm" w={tokens.layout.nav} bg="var(--sigta-nav-background)">
      <UserMenu user={staff} onLogout={() => {}} />
    </Paper>
  ),
  parameters: { layout: 'centered' },
}
