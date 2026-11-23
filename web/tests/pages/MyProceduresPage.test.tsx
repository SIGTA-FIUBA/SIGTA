import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { render } from '../test-utils/render'
import { MyProceduresPage } from '../../src/pages/MyProceduresPage'

const inOffice = {
  id: 'P-1',
  procedure_type: 1,
  procedure_type_name: 'Equivalencia entre asignaturas',
  step_name: 'Revisión de la solicitud',
  state: 'in_office',
  area: { id: 'secretariat', name: 'Secretaría' },
  correction_pending: false,
  started_at: '2026-11-02T12:00:00Z',
  last_moved_at: '2026-11-10T15:00:00Z',
  resolutions: [],
}
const closed = {
  ...inOffice,
  id: 'P-3',
  procedure_type_name: 'Pase de carrera',
  state: 'closed',
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

// stubApi plays the API: /me/procedures answers the procedures the person
// started.
function stubApi(procedures: unknown[], status = 200) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => (status === 200 ? Response.json(procedures) : Response.json({ message: 'error' }, { status }))),
  )
}

function renderPage() {
  const router = createMemoryRouter([{ path: '/my-procedures', element: <MyProceduresPage /> }], {
    initialEntries: ['/my-procedures'],
  })
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

describe('MyProceduresPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('lists the procedures in progress, each with its state and leading to its detail', async () => {
    stubApi([inOffice, withApplicant])
    renderPage()

    const list = await screen.findByRole('list', { name: 'Trámites en curso' })
    const first = within(list).getByRole('link', { name: /Equivalencia entre asignaturas/ })
    expect(first).toHaveAttribute('href', '/procedures/P-1')
    expect(within(first).getByText('En Secretaría')).toBeVisible()
    const second = within(list).getByRole('link', { name: /Reconocimiento de créditos/ })
    expect(within(second).getByText('Acción requerida')).toBeVisible()
    expect(screen.getByRole('heading', { name: 'En curso (2)' })).toBeVisible()
    expect(screen.getByRole('heading', { level: 1, name: 'Historial' })).toBeVisible()
  })

  it('lists the closed procedures apart, with their resolution', async () => {
    stubApi([inOffice, closed])
    renderPage()

    const closedOnes = await screen.findByRole('list', { name: 'Trámites cerrados' })
    const row = within(closedOnes).getByRole('link', { name: /Pase de carrera/ })
    expect(within(row).getByText('Cierre: 05/11/2026')).toBeVisible()
    expect(screen.getByRole('link', { name: 'Descargar resolucion.pdf de Pase de carrera' })).toHaveAttribute(
      'href',
      '/api/procedures/P-3/documents/D-9/content',
    )
    const inProgress = screen.getByRole('list', { name: 'Trámites en curso' })
    expect(within(inProgress).queryByRole('link', { name: /Pase de carrera/ })).not.toBeInTheDocument()
  })

  it('says when a closed procedure has no resolution', async () => {
    stubApi([{ ...closed, resolutions: [] }])
    renderPage()

    expect(await screen.findByText('El trámite se cerró sin un documento resolutivo adjunto.')).toBeVisible()
  })

  it('narrows the procedures by name, with or without accents', async () => {
    stubApi([inOffice, withApplicant])
    renderPage()

    await screen.findByRole('list', { name: 'Trámites en curso' })
    await userEvent.type(screen.getByRole('textbox', { name: 'Buscar por nombre' }), 'creditos')
    const list = screen.getByRole('list', { name: 'Trámites en curso' })
    expect(within(list).getAllByRole('link')).toHaveLength(1)
    expect(within(list).getByRole('link', { name: /Reconocimiento de créditos/ })).toBeVisible()
  })

  it('says when no procedure was started', async () => {
    stubApi([])
    renderPage()

    expect(await screen.findByText('Sin trámites')).toBeVisible()
  })

  it('says when the procedures could not be loaded', async () => {
    stubApi([], 500)
    renderPage()

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudieron cargar los trámites')
  })
})
