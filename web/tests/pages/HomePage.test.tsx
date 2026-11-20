import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import { createMemoryRouter, Outlet, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { render } from '../test-utils/render'
import { HomePage } from '../../src/pages/HomePage'

const inOffice = {
  id: 'P-1',
  procedure_type: 1,
  procedure_type_name: 'Equivalencia entre asignaturas',
  step_name: 'Revisión de la solicitud',
  state: 'in_office',
  started_at: '2026-11-02T12:00:00Z',
  last_moved_at: '2026-11-10T15:00:00Z',
}
const withApplicant = {
  ...inOffice,
  id: 'P-2',
  procedure_type_name: 'Reconocimiento de créditos',
  step_name: 'Documentación',
  state: 'with_applicant',
}

function stubApi(status: number, body: unknown) {
  const fetch = vi.fn(async (_request: Request) => Response.json(body, { status }))
  vi.stubGlobal('fetch', fetch)
  return fetch
}

function renderAs(staff: boolean) {
  const router = createMemoryRouter([
    {
      element: <Outlet context={{ id: 'u1', email: 'u@fi.uba.ar', name: 'Persona', staff, administrator: false }} />,
      children: [{ path: '/', element: <HomePage /> }],
    },
  ])
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('HomePage', () => {
  it('says what the platform is', () => {
    stubApi(200, [])
    renderAs(false)

    expect(
      screen.getByRole('heading', { level: 1, name: 'Sistema Integral de Gestión de Trámites Académicos' }),
    ).toBeVisible()
    expect(screen.getByRole('heading', { name: 'Criterios de la plataforma' })).toBeVisible()
  })

  it('lists the procedures in progress of the person, each with its state and leading to its detail', async () => {
    const fetch = stubApi(200, [inOffice, withApplicant])
    renderAs(false)

    const mine = screen.getByRole('region', { name: 'Mis trámites en curso' })
    const first = await within(mine).findByRole('link', { name: /Equivalencia entre asignaturas/ })
    expect(first).toHaveAttribute('href', '/procedures/P-1')
    expect(within(first).getByText('En un área')).toBeVisible()
    expect(within(first).getByText('Paso actual: Revisión de la solicitud')).toBeVisible()
    expect(within(first).getByText('Último movimiento: 10/11/2026')).toBeVisible()
    const second = within(mine).getByRole('link', { name: /Reconocimiento de créditos/ })
    expect(within(second).getByText('Pendiente del titular')).toBeVisible()
    expect(new URL(fetch.mock.calls[0]![0].url).pathname).toBe('/api/me/procedures')
  })

  it('says when there is no procedure in progress', async () => {
    stubApi(200, [])
    renderAs(true)

    expect(await screen.findByText(/No hay trámites en curso/)).toBeVisible()
  })

  it('says when the procedures in progress could not be loaded', async () => {
    stubApi(500, { message: 'error' })
    renderAs(false)

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudieron cargar los trámites en curso')
  })

  it('explains to an applicant how a procedure is done, and links to the catalog', () => {
    stubApi(200, [])
    renderAs(false)

    expect(screen.getByRole('heading', { name: 'Cómo se realiza un trámite' })).toBeVisible()
    expect(screen.getByRole('link', { name: /Catálogo de trámites/ })).toHaveAttribute('href', '/catalog')
    expect(screen.queryByRole('link', { name: /Bandeja de trámites/ })).not.toBeInTheDocument()
  })

  it('explains to the staff how to work, and links to the inbox', () => {
    stubApi(200, [])
    renderAs(true)

    expect(screen.getByRole('heading', { name: 'Cómo se trabaja en SIGTA' })).toBeVisible()
    expect(screen.getByRole('link', { name: /Bandeja de trámites/ })).toHaveAttribute('href', '/inbox')
  })
})
