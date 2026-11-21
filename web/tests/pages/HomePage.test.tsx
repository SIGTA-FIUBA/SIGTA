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
const closed = {
  id: 'P-3',
  procedure_type: 1,
  procedure_type_name: 'Pase de carrera',
  state: 'closed',
  started_at: '2026-10-01T12:00:00Z',
  last_moved_at: '2026-11-05T15:00:00Z',
  resolutions: [
    {
      id: 'D-9',
      document: 'dean_resolution',
      document_name: 'Resolución de Decano',
      file_name: 'resolucion.pdf',
      media_type: 'application/pdf',
      size: 2048,
    },
  ],
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

  it('lists the closed procedures apart, with the date they closed and leading to their detail', async () => {
    stubApi(200, [inOffice, closed])
    renderAs(false)

    const closedOnes = await screen.findByRole('region', { name: 'Mis trámites cerrados' })
    expect(within(closedOnes).getByRole('link', { name: 'Pase de carrera' })).toHaveAttribute('href', '/procedures/P-3')
    expect(within(closedOnes).getByText('Cierre: 05/11/2026')).toBeVisible()
    expect(within(closedOnes).queryByText(/Paso actual/)).not.toBeInTheDocument()
    const inProgress = screen.getByRole('region', { name: 'Mis trámites en curso' })
    expect(within(inProgress).queryByRole('link', { name: /Pase de carrera/ })).not.toBeInTheDocument()
  })

  it('offers the resolution of a closed procedure to download without opening its detail', async () => {
    stubApi(200, [closed])
    renderAs(false)

    const closedOnes = await screen.findByRole('region', { name: 'Mis trámites cerrados' })
    expect(within(closedOnes).getByText('Resolución de Decano')).toBeVisible()
    expect(within(closedOnes).getByRole('link', { name: 'Descargar resolucion.pdf' })).toHaveAttribute(
      'href',
      '/api/procedures/P-3/documents/D-9/content',
    )
    expect(within(closedOnes).getByRole('button', { name: 'Ver resolucion.pdf' })).toBeVisible()
  })

  it('says when a closed procedure has no resolution', async () => {
    stubApi(200, [{ ...closed, resolutions: [] }])
    renderAs(false)

    const closedOnes = await screen.findByRole('region', { name: 'Mis trámites cerrados' })
    expect(within(closedOnes).getByText('El trámite se cerró sin un documento resolutivo adjunto.')).toBeVisible()
    expect(within(closedOnes).queryByRole('link', { name: /Descargar/ })).not.toBeInTheDocument()
  })

  it('says there is no procedure in progress when all of them are closed', async () => {
    stubApi(200, [closed])
    renderAs(false)

    expect(await screen.findByText(/No hay trámites en curso/)).toBeVisible()
    expect(screen.getByRole('region', { name: 'Mis trámites cerrados' })).toBeVisible()
  })

  it('leaves out the closed section when no procedure is closed', async () => {
    stubApi(200, [inOffice])
    renderAs(false)

    await screen.findByRole('link', { name: /Equivalencia entre asignaturas/ })
    expect(screen.queryByRole('region', { name: 'Mis trámites cerrados' })).not.toBeInTheDocument()
  })

  it('says when the applicant has a correction to make', async () => {
    stubApi(200, [{ ...withApplicant, correction_pending: true }, inOffice])
    renderAs(false)

    const pending = await screen.findByRole('link', { name: /Reconocimiento de créditos/ })
    expect(within(pending).getByText('Subsanación pendiente del titular')).toBeVisible()
    const other = screen.getByRole('link', { name: /Equivalencia entre asignaturas/ })
    expect(within(other).queryByText(/Subsanación pendiente/)).toBeNull()
  })

  it('says when there is no procedure in progress', async () => {
    stubApi(200, [])
    renderAs(true)

    expect(await screen.findByText(/No hay trámites en curso/)).toBeVisible()
  })

  it('says when the procedures in progress could not be loaded', async () => {
    stubApi(500, { message: 'error' })
    renderAs(false)

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudieron cargar los trámites')
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
