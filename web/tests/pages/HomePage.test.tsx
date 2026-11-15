import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import { createMemoryRouter, Outlet, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { render } from '../render'
import { HomePage } from '../../src/pages/HomePage'

const item = {
  procedure_type: '2',
  procedure_type_name: 'Equivalencia',
  applicant_name: 'Titular Uno',
  step_name: 'Registrar expediente',
  area: 'office',
  area_name: 'Dirección de Alumnos',
  state: 'in_office',
  actionable: true,
  started_at: '2026-11-01T10:00:00Z',
  last_moved_at: '2026-11-03T10:00:00Z',
  age_days: 3,
}

function stubProcedures(items: unknown[]) {
  const fetch = vi.fn(async (_request: Request) => Response.json({ items, procedure_types: [], areas: [] }))
  vi.stubGlobal('fetch', fetch)
  return fetch
}

function renderAs(role: string) {
  const router = createMemoryRouter([
    {
      element: <Outlet context={{ id: 'u1', email: 'u@fi.uba.ar', name: 'Persona', role }} />,
      children: [{ path: '/', element: <HomePage /> }],
    },
  ])
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

describe('HomePage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('offers an applicant to start a procedure, and how it is done', () => {
    renderAs('applicant')

    expect(screen.getByRole('heading', { level: 1, name: 'Inicio' })).toBeVisible()
    expect(screen.getByRole('link', { name: 'Iniciar un trámite' })).toHaveAttribute('href', '/catalog')
    expect(screen.getByRole('heading', { name: 'Cómo se realiza un trámite' })).toBeVisible()
    expect(screen.queryByRole('heading', { name: 'Requieren acción' })).not.toBeInTheDocument()
  })

  it('lists for the staff the procedures that need action, the latest news first', async () => {
    stubProcedures([
      { ...item, id: 'old', applicant_name: 'Titular Uno' },
      { ...item, id: 'new', applicant_name: 'Titular Dos', last_moved_at: '2026-11-05T10:00:00Z' },
      { ...item, id: 'waiting', applicant_name: 'Titular Tres', actionable: false },
    ])
    renderAs('staff_operative')

    expect(screen.getByRole('link', { name: 'Ir a la bandeja' })).toHaveAttribute('href', '/inbox')
    const list = await screen.findByRole('list', { name: 'Trámites que requieren acción' })
    const links = within(list).getAllByRole('link')
    expect(links.map((link) => link.getAttribute('href'))).toEqual(['/procedures/new', '/procedures/old'])
    expect(within(links[0]!).getByText('Titular Dos')).toBeVisible()
    expect(within(links[0]!).getByText('En la Secretaría')).toBeVisible()
    expect(within(links[0]!).getByText('Próximo paso: Registrar expediente')).toBeVisible()
  })

  it('says so when no procedure needs action', async () => {
    stubProcedures([{ ...item, id: 'waiting', actionable: false }])
    renderAs('staff_operative')

    expect(await screen.findByText('No hay trámites pendientes')).toBeVisible()
  })

  it('shows the staff no steps, only what needs action', () => {
    stubProcedures([])
    renderAs('staff_operative')

    expect(screen.queryByRole('heading', { name: 'Cómo se realiza un trámite' })).not.toBeInTheDocument()
  })
})
