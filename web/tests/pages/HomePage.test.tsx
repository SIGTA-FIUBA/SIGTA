import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import { createMemoryRouter, Outlet, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { render } from '../test-utils/render'
import { HomePage } from '../../src/pages/HomePage'

const item = {
  procedure_type: 2,
  procedure_type_name: 'Equivalencia',
  applicant: { id: 'a1', name: 'Titular Uno' },
  step_name: 'Registrar expediente',
  area: { id: 'office', name: 'Dirección de Alumnos' },
  state: 'in_office',
  actionable: true,
  started_at: '2026-11-01T10:00:00Z',
  last_moved_at: '2026-11-03T10:00:00Z',
  age_days: 3,
}

const inOffice = {
  id: 'P-1',
  procedure_type: 1,
  procedure_type_name: 'Equivalencia entre asignaturas',
  step_name: 'Revisión de la solicitud',
  state: 'in_office',
  correction_pending: false,
  started_at: '2026-11-02T12:00:00Z',
  last_moved_at: '2026-11-10T15:00:00Z',
}
const closed = {
  id: 'P-3',
  procedure_type: 1,
  procedure_type_name: 'Pase de carrera',
  state: 'closed',
  correction_pending: false,
  started_at: '2026-10-01T12:00:00Z',
  last_moved_at: '2026-11-05T15:00:00Z',
}
const withApplicant = {
  ...inOffice,
  id: 'P-2',
  procedure_type_name: 'Reconocimiento de créditos',
  step_name: 'Documentación',
  state: 'with_applicant',
}

// stubApi plays the API: /me/procedures answers the procedures the person
// started, and /procedures, the inbox of the staff.
function stubApi({ mine = [], inbox = [], status = 200 }: { mine?: unknown[]; inbox?: unknown[]; status?: number }) {
  const fetch = vi.fn(async (request: Request) => {
    if (status !== 200) {
      return Response.json({ message: 'error' }, { status })
    }
    if (new URL(request.url).pathname === '/api/me/procedures') {
      return Response.json(mine)
    }
    return Response.json({ items: inbox, procedure_types: [], areas: [] })
  })
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

  it('greets an applicant and shows how a procedure is done, with no button that repeats the menu', () => {
    stubApi({})
    renderAs(false)

    expect(screen.getByRole('heading', { level: 1, name: 'Hola, Persona' })).toBeVisible()
    expect(screen.queryByRole('link', { name: 'Iniciar un trámite' })).not.toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Cómo se realiza un trámite' })).toBeVisible()
    expect(screen.queryByRole('heading', { name: 'Requieren acción' })).not.toBeInTheDocument()
  })

  it('lists the procedures in progress of the person, each with its state and leading to its detail', async () => {
    const fetch = stubApi({ mine: [inOffice, withApplicant] })
    renderAs(false)

    const mine = await screen.findByRole('list', { name: 'Mis trámites en curso' })
    const first = within(mine).getByRole('link', { name: /Equivalencia entre asignaturas/ })
    expect(first).toHaveAttribute('href', '/procedures/P-1')
    expect(within(first).getByText('En un área')).toBeVisible()
    expect(within(first).getByText('Revisión de la solicitud')).toBeVisible()
    const second = within(mine).getByRole('link', { name: /Reconocimiento de créditos/ })
    expect(within(second).getByText('Pendiente del titular')).toBeVisible()
    expect(new URL(fetch.mock.calls[0]![0].url).pathname).toBe('/api/me/procedures')
  })

  it('lists the closed procedures apart, leading to their detail', async () => {
    stubApi({ mine: [inOffice, closed] })
    renderAs(false)

    const closedOnes = await screen.findByRole('list', { name: 'Mis trámites cerrados' })
    const row = within(closedOnes).getByRole('link', { name: /Pase de carrera/ })
    expect(row).toHaveAttribute('href', '/procedures/P-3')
    expect(within(row).getByText('Cierre: 05/11/2026')).toBeVisible()
    expect(within(row).getByText('Cerrado')).toBeVisible()
    const inProgress = screen.getByRole('list', { name: 'Mis trámites en curso' })
    expect(within(inProgress).queryByRole('link', { name: /Pase de carrera/ })).not.toBeInTheDocument()
  })

  it('says there is no procedure in progress when all of them are closed', async () => {
    stubApi({ mine: [closed] })
    renderAs(false)

    expect(await screen.findByText(/No hay trámites en curso/)).toBeVisible()
    expect(screen.getByRole('region', { name: 'Mis trámites cerrados' })).toBeVisible()
  })

  it('leaves out the closed section when no procedure is closed', async () => {
    stubApi({ mine: [inOffice] })
    renderAs(false)

    await screen.findByRole('link', { name: /Equivalencia entre asignaturas/ })
    expect(screen.queryByRole('region', { name: 'Mis trámites cerrados' })).not.toBeInTheDocument()
  })

  it('says when the applicant has a correction to make', async () => {
    stubApi({ mine: [{ ...withApplicant, correction_pending: true }, inOffice] })
    renderAs(false)

    const pending = await screen.findByRole('link', { name: /Reconocimiento de créditos/ })
    expect(within(pending).getByText('Subsanación pendiente del titular')).toBeVisible()
    const other = screen.getByRole('link', { name: /Equivalencia entre asignaturas/ })
    expect(within(other).queryByText(/Subsanación pendiente/)).toBeNull()
  })

  it('says when the procedures in progress could not be loaded', async () => {
    stubApi({ status: 500 })
    renderAs(false)

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudieron cargar los trámites')
  })

  it('lists for the staff the procedures that need action, the latest news first', async () => {
    stubApi({
      inbox: [
        { ...item, id: 'old' },
        { ...item, id: 'new', applicant: { id: 'a2', name: 'Titular Dos' }, last_moved_at: '2026-11-05T10:00:00Z' },
        { ...item, id: 'waiting', applicant: { id: 'a3', name: 'Titular Tres' }, actionable: false },
      ],
    })
    renderAs(true)

    expect(screen.queryByRole('link', { name: 'Ir a la bandeja' })).not.toBeInTheDocument()
    const list = await screen.findByRole('list', { name: 'Trámites que requieren acción' })
    const links = within(list).getAllByRole('link')
    expect(links.map((link) => link.getAttribute('href'))).toEqual(['/procedures/new', '/procedures/old'])
    expect(within(links[0]!).getByText('Equivalencia')).toBeVisible()
    expect(within(links[0]!).getByText('Titular Dos')).toBeVisible()
    expect(within(links[0]!).getByText('Registrar expediente')).toBeVisible()
    expect(within(links[0]!).getByText('En un área')).toBeVisible()
    expect(within(links[0]!).getByText('A cargo de Dirección de Alumnos')).toBeVisible()
    expect(screen.getByRole('link', { name: 'Ver todos' })).toHaveAttribute('href', '/inbox')
  })

  it('shows at most five procedures that need action', async () => {
    stubApi({ inbox: Array.from({ length: 7 }, (_, index) => ({ ...item, id: `p${index}` })) })
    renderAs(true)

    const list = await screen.findByRole('list', { name: 'Trámites que requieren acción' })
    expect(within(list).getAllByRole('link')).toHaveLength(5)
  })

  it('sums up what needs action and what is in progress', async () => {
    stubApi({
      inbox: [
        { ...item, id: 'a' },
        { ...item, id: 'b', actionable: false },
      ],
    })
    renderAs(true)

    const summary = await screen.findByLabelText('Resumen')
    expect(within(summary).getByRole('link', { name: '1 Requieren acción' })).toHaveAttribute('href', '/inbox')
    expect(within(summary).getByRole('link', { name: '2 En curso' })).toHaveAttribute('href', '/inbox?view=all')
  })

  it('lists the latest moves, one line each', async () => {
    stubApi({
      inbox: [
        { ...item, id: 'a', actionable: false },
        { ...item, id: 'b', applicant: { id: 'a2', name: 'Titular Dos' }, last_moved_at: '2026-11-05T10:00:00Z' },
      ],
    })
    renderAs(true)

    const activity = await screen.findByRole('list', { name: 'Movimientos recientes' })
    const lines = within(activity).getAllByRole('listitem')
    expect(lines[0]).toHaveTextContent('Equivalencia de Titular Dos pasó a Registrar expediente')
    expect(lines[1]).toHaveTextContent('Equivalencia de Titular Uno pasó a Registrar expediente')
  })

  it('says so when no procedure needs action', async () => {
    stubApi({ inbox: [{ ...item, id: 'waiting', actionable: false }] })
    renderAs(true)

    expect(await screen.findByText('No hay trámites pendientes')).toBeVisible()
  })

  it('shows the staff their own procedures too, and no steps', async () => {
    stubApi({ mine: [inOffice] })
    renderAs(true)

    expect(await screen.findByRole('link', { name: /Equivalencia entre asignaturas/ })).toBeVisible()
    expect(screen.queryByRole('heading', { name: 'Cómo se realiza un trámite' })).not.toBeInTheDocument()
  })
})
