import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen } from '@testing-library/react'
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
  it('leaves for the identity provider through the API', () => {
    renderAt('/login')

    expect(screen.getByRole('link', { name: 'Continuar con Google' })).toHaveAttribute('href', '/api/auth/login')
    expect(screen.getByRole('link', { name: 'Ingresar' })).toHaveAttribute('href', '/api/auth/login')
    expect(screen.getByText('Con tu cuenta institucional de la UBA.')).toBeVisible()
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('presents the platform and how a procedure is done', () => {
    renderAt('/login')

    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent(
      'Tus trámites académicos, en línea y con seguimiento',
    )
    expect(screen.getByRole('heading', { name: 'Un solo lugar, del inicio a la resolución' })).toBeVisible()
    expect(screen.getByRole('heading', { name: 'Seguimiento' })).toBeVisible()
    expect(screen.getByRole('heading', { name: 'Las decisiones son de una persona' })).toBeVisible()
    expect(screen.getByRole('link', { name: 'Ver el catálogo' })).toHaveAttribute('href', '/catalog')
  })

  it('counts the catalog and shows its first procedures in order', async () => {
    const names = ['Pase', 'Equivalencia', 'Reconocimiento', 'Intercambio', 'Baja']
    renderAt('/login', { status: 200, body: names.map((name, index) => catalogEntry(index + 1, name)) })

    expect(await screen.findByText('5')).toBeVisible()
    expect(screen.getByText('trámites disponibles, con sus requisitos publicados')).toBeVisible()
    expect(screen.getByRole('link', { name: /Baja/ })).toHaveAttribute('href', '/catalog')
    expect(screen.getByRole('link', { name: /Pase/ })).toBeVisible()
    expect(screen.queryByRole('link', { name: /Reconocimiento/ })).not.toBeInTheDocument()
  })

  it('says it when the catalog has a single procedure', async () => {
    renderAt('/login', { status: 200, body: [catalogEntry(1, 'Equivalencia')] })

    expect(await screen.findByText('trámite disponible, con sus requisitos publicados')).toBeVisible()
  })

  it('says it when the catalog is still empty', async () => {
    renderAt('/login')

    expect(await screen.findByText('Todavía no hay trámites publicados.')).toBeVisible()
  })

  it('says it when the catalog does not load', async () => {
    renderAt('/login', { status: 500, body: {} })

    expect(await screen.findByText('No pudimos cargar los trámites. Volvé a intentarlo más tarde.')).toBeVisible()
    expect(screen.getByRole('link', { name: 'Ver el catálogo completo' })).toHaveAttribute('href', '/catalog')
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
