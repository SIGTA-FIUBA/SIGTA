import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { InboxPage } from '../../src/pages/InboxPage'
import { ProcedurePage } from '../../src/pages/ProcedurePage'
import { render } from '../render'

const day = 86_400_000
const daysAgo = (days: number) => new Date(Date.now() - days * day - 3_600_000).toISOString()

// pending: a step of an area this account can do. evaluating: in another
// area, which this account does not act on. waiting: on the applicant, so
// with no area. Each of a different type or area.
const pending = {
  id: 'p-1',
  procedure_type: 2,
  procedure_type_name: 'Equivalencia',
  applicant: { id: 'a-1', name: 'Titular Uno' },
  step_name: 'Registrar expediente',
  area: { id: 'office', name: 'Dirección de Alumnos' },
  state: 'in_office',
  actionable: true,
  started_at: daysAgo(3),
  last_moved_at: daysAgo(3),
  age_days: 3,
}
const evaluating = {
  ...pending,
  id: 'p-2',
  applicant: { id: 'a-2', name: 'Titular Dos' },
  step_name: 'Dictamen',
  area: { id: 'committee', name: 'Comisión Curricular' },
  actionable: false,
  started_at: daysAgo(30),
  last_moved_at: daysAgo(12),
  age_days: 30,
}
const waiting = {
  ...pending,
  id: 'p-3',
  procedure_type: 3,
  procedure_type_name: 'Créditos',
  applicant: { id: 'a-3', name: 'Titular Tres' },
  step_name: 'Documentación pedida',
  area: undefined,
  state: 'with_applicant',
  actionable: false,
  started_at: daysAgo(8),
  last_moved_at: daysAgo(1),
  age_days: 8,
}

const options = {
  procedure_types: [
    { id: '3', name: 'Créditos' },
    { id: '2', name: 'Equivalencia' },
  ],
  areas: [
    { id: 'committee', name: 'Comisión Curricular' },
    { id: 'office', name: 'Dirección de Alumnos' },
  ],
}

function stubApi(status: number, body: unknown) {
  const fetch = vi.fn(async (_request: Request) => Response.json(body, { status }))
  vi.stubGlobal('fetch', fetch)
  return fetch
}

function renderAt(path = '/inbox') {
  const router = createMemoryRouter(
    [
      { path: '/inbox', element: <InboxPage /> },
      { path: '/procedures/:id', element: <ProcedurePage /> },
    ],
    { initialEntries: [path] },
  )
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
  return router
}

// The table and the cards are both in the DOM: CSS shows one per screen width.
// The tests read the table.
async function tableRows() {
  const table = await screen.findByRole('table')
  return within(table).getAllByRole('row').slice(1)
}

function listedNames() {
  const table = screen.queryByRole('table')
  if (table === null) {
    return []
  }
  return within(table)
    .getAllByRole('link')
    .map((link) => link.textContent)
}

describe('InboxPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('asks once for everything the person can see', async () => {
    const fetch = stubApi(200, { items: [pending], ...options })
    renderAt()

    await tableRows()
    expect(fetch).toHaveBeenCalledTimes(1)
    expect(Object.fromEntries(new URL(fetch.mock.calls[0]![0].url).searchParams)).toEqual({ view: 'all' })
  })

  it('shows the procedure type, pending step, who has it, state and age', async () => {
    stubApi(200, { items: [pending], ...options })
    renderAt()

    const [row] = await tableRows()
    expect(within(row!).getByRole('link', { name: 'Titular Uno' })).toBeVisible()
    expect(within(row!).getByText('Equivalencia')).toBeVisible()
    expect(within(row!).getByText('Registrar expediente')).toBeVisible()
    expect(within(row!).getByText('Dirección de Alumnos')).toBeVisible()
    expect(within(row!).getByText('En un área')).toBeVisible()
    expect(within(row!).getByText('3 días')).toBeVisible()
    expect(within(row!).getByText('Hace 3 días')).toBeVisible()
  })

  it('shows the applicant in charge of a step of their own, and a name the API does not know', async () => {
    stubApi(200, { items: [{ ...waiting, applicant: { id: 'a-3', name: '' } }], ...options })
    renderAt('/inbox?view=all')

    const [row] = await tableRows()
    expect(within(row!).getByRole('link', { name: 'Titular sin nombre registrado' })).toBeVisible()
    expect(within(row!).getByText('Titular')).toBeVisible()
  })

  it('shows only what needs action by default', async () => {
    stubApi(200, { items: [pending, evaluating, waiting], ...options })
    renderAt()

    await tableRows()
    expect(listedNames()).toEqual(['Titular Uno'])
    expect(screen.getByRole('radio', { name: 'Asignados (1)' })).toBeChecked()
    expect(screen.getByRole('radio', { name: 'Todos (3)' })).not.toBeChecked()
  })

  it('counts each state over everything visible, whatever the filters', async () => {
    stubApi(200, { items: [pending, evaluating, waiting], ...options })
    renderAt('/inbox?view=all&type=3')

    await tableRows()
    expect(screen.getByRole('button', { name: /Acción requerida\s*1/ })).toBeVisible()
    expect(screen.getByRole('button', { name: /En un área\s*2/ })).toBeVisible()
    expect(screen.getByRole('button', { name: /Pendiente del titular\s*1/ })).toBeVisible()
  })

  it('lists everything in progress, marking what needs action, oldest first', async () => {
    stubApi(200, { items: [pending, evaluating, waiting], ...options })
    renderAt('/inbox?view=all')

    const rows = await tableRows()
    expect(listedNames()).toEqual(['Titular Dos', 'Titular Tres', 'Titular Uno'])
    expect(within(rows[2]!).getByText('Acción requerida')).toBeVisible()
    expect(within(rows[0]!).queryByText('Acción requerida')).not.toBeInTheDocument()
  })

  it('filters by state from a summary and by the selects, combined', async () => {
    stubApi(200, { items: [pending, evaluating, waiting], ...options })
    const router = renderAt()
    await tableRows()

    await userEvent.click(screen.getByRole('button', { name: /En un área/ }))
    expect(listedNames()).toEqual(['Titular Dos', 'Titular Uno'])

    await userEvent.click(screen.getByRole('combobox', { name: 'Tipo de trámite' }))
    await userEvent.click(screen.getByRole('option', { name: 'Créditos' }))
    expect(listedNames()).toEqual([])
    expect(screen.getByText('No hay trámites que coincidan con los filtros.')).toBeVisible()
    expect(Object.fromEntries(new URLSearchParams(router.state.location.search))).toEqual({
      view: 'all',
      state: 'in_office',
      type: '3',
    })
  })

  it('sorts by the last movement', async () => {
    stubApi(200, { items: [pending, evaluating, waiting], ...options })
    renderAt('/inbox?view=all&sort=last_moved')

    await tableRows()
    expect(listedNames()).toEqual(['Titular Tres', 'Titular Uno', 'Titular Dos'])
  })

  it('keeps the filters when coming back from a procedure', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (request: Request) =>
        new URL(request.url).pathname.endsWith('/procedures/p-1')
          ? Response.json({ message: 'No existe el trámite.' }, { status: 404 })
          : Response.json({ items: [pending, evaluating, waiting], ...options }),
      ),
    )
    const router = renderAt('/inbox?view=all&area=office')

    const [row] = await tableRows()
    await userEvent.click(within(row!).getByRole('link', { name: 'Titular Uno' }))
    expect(router.state.location.pathname).toBe('/procedures/p-1')
    await userEvent.click(screen.getByRole('link', { name: 'Volver a la bandeja de trámites' }))

    await tableRows()
    expect(router.state.location.search).toBe('?view=all&area=office')
    expect(listedNames()).toEqual(['Titular Uno'])
  })

  it('says when nothing needs action', async () => {
    stubApi(200, { items: [evaluating], ...options })
    renderAt()

    expect(await screen.findByText('No hay trámites que requieran acción de esta cuenta.')).toBeVisible()
  })

  it('says when there is nothing in progress', async () => {
    stubApi(200, { items: [], procedure_types: [], areas: [] })
    renderAt('/inbox?view=all')

    expect(await screen.findByText('No hay trámites en curso.')).toBeVisible()
  })

  it('tells who is not staff that the inbox is not theirs', async () => {
    stubApi(403, { message: 'La bandeja es solo para el personal.' })
    renderAt()

    expect(await screen.findByRole('alert')).toHaveTextContent('disponible solo para el personal')
  })

  it('says when the inbox fails', async () => {
    stubApi(500, { message: 'Ocurrió un error inesperado.' })
    renderAt()

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudo cargar la bandeja de trámites.')
  })

  it('ignores filters in the URL that it does not know', async () => {
    stubApi(200, { items: [pending, evaluating], ...options })
    renderAt('/inbox?view=everything&state=approved&type=abc&sort=name')

    await tableRows()
    expect(listedNames()).toEqual(['Titular Uno'])
  })
})
