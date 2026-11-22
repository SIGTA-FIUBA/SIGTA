import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, waitFor, within } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { render } from '../test-utils/render'
import { LoginPage } from '../../src/pages/LoginPage'

function catalogEntry(id: number, name: string) {
  return {
    id,
    version: 1,
    name,
    description: `Descripción de ${name}`,
    deadlines: [],
    admissibility: [],
    fields: [],
    documents: [],
    submission: { fields: [], documents: [] },
  }
}

function renderAt(path: string, catalog: { status: number; body: unknown } = { status: 200, body: [] }) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => Response.json(catalog.body, { status: catalog.status })),
  )
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const router = createMemoryRouter([{ path: '/login', element: <LoginPage /> }], {
    initialEntries: [path],
  })
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('LoginPage', () => {
  it('leaves for the identity provider through the API', async () => {
    renderAt('/login')

    expect(screen.getByRole('link', { name: 'Ingresar a SIGTA' })).toHaveAttribute('href', '/api/auth/login')
    expect(screen.getByRole('link', { name: 'Ingresar' })).toHaveAttribute('href', '/api/auth/login')
    // The first screen arrives with the page, in a frame.
    await waitFor(() => expect(screen.getByText('Con tu cuenta institucional de la UBA.')).toBeVisible())
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('opens with what a visitor can do and leads to the requirements', () => {
    renderAt('/login')

    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Tus trámites académicos')
    expect(screen.getByRole('link', { name: 'Consultar requisitos' })).toHaveAttribute('href', '/catalog')
    expect(
      screen.getByRole('img', { name: 'Dos estudiantes repasan un apunte en un pasillo de la facultad.' }),
    ).toBeInTheDocument()
    // What SIGTA is and its criteria are in "Acerca de SIGTA", not here.
    expect(screen.queryByRole('heading', { name: 'Cómo trabaja la plataforma' })).not.toBeInTheDocument()
  })

  it('lists the available procedures by name, in alphabetical order', async () => {
    const names = ['Readmisión', 'Equivalencia de asignaturas', 'Constancia de alumno regular']
    renderAt('/login', { status: 200, body: names.map((name, index) => catalogEntry(index + 1, name)) })

    const list = await screen.findByRole('list')
    expect(
      within(list)
        .getAllByRole('listitem')
        .map((item) => item.firstChild?.textContent),
    ).toEqual(['Constancia de alumno regular', 'Equivalencia de asignaturas', 'Readmisión'])
    expect(screen.getByText('Descripción de Readmisión')).toBeVisible()
  })

  it('says it once when the catalog is still empty', async () => {
    renderAt('/login')

    expect(await screen.findByText('Todavía no hay trámites publicados.')).toBeVisible()
    expect(screen.queryByRole('list')).not.toBeInTheDocument()
  })

  it('says it when the catalog does not load', async () => {
    renderAt('/login', { status: 500, body: {} })

    expect(await screen.findByText('No se pudieron cargar los trámites. Volvé a intentarlo más tarde.')).toBeVisible()
  })

  it.each([
    ['state_mismatch', 'El ingreso no coincide con el que se inició. Volvé a intentarlo.'],
    ['rejected', 'El ingreso no se pudo completar. Volvé a intentarlo.'],
    ['domain_not_allowed', 'Ingresá con tu cuenta institucional.'],
  ])('explains the %s error', (error, message) => {
    renderAt(`/login?error=${error}`)

    expect(screen.getByRole('alert')).toHaveTextContent(message)
  })

  it('treats an unknown error as a rejected login', () => {
    renderAt('/login?error=constructor')

    expect(screen.getByRole('alert')).toHaveTextContent('El ingreso no se pudo completar.')
  })

  it('signs with FIUBA in the top bar and in the footer', () => {
    renderAt('/login')

    expect(screen.getAllByRole('img', { name: 'Facultad de Ingeniería, Universidad de Buenos Aires' })).toHaveLength(2)
  })
})
