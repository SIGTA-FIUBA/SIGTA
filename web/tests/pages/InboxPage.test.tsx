import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, Outlet, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { InboxPage } from '../../src/pages/InboxPage'
import { ProcedurePage } from '../../src/pages/ProcedurePage'
import { render } from '../render'

const day = 86_400_000
const daysAgo = (days: number) => new Date(Date.now() - days * day - 3_600_000).toISOString()

// pending: an office step this profile can do. referred: with an external
// evaluator. waiting: on the applicant. Each of a different type and area.
const pending = {
  id: 'p-1',
  procedure_type: 2,
  procedure_type_name: 'Equivalencia',
  applicant_name: 'Titular Uno',
  step_name: 'Registrar expediente',
  area: 'office',
  area_name: 'Dirección de Alumnos',
  state: 'in_office',
  actionable: true,
  started_at: daysAgo(3),
  last_moved_at: daysAgo(3),
  age_days: 3,
}
const referred = {
  ...pending,
  id: 'p-2',
  applicant_name: 'Titular Dos',
  step_name: 'Dictamen',
  area: 'committee',
  area_name: 'Comisión Curricular',
  state: 'referred',
  actionable: false,
  started_at: daysAgo(30),
  last_moved_at: daysAgo(12),
  age_days: 30,
  days_referred: 12,
}
const waiting = {
  ...pending,
  id: 'p-3',
  procedure_type: 3,
  procedure_type_name: 'Créditos',
  applicant_name: 'Titular Tres',
  step_name: 'Documentación pedida',
  area: 'applicant',
  area_name: 'Titular',
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
    { id: 'applicant', name: 'Titular' },
  ],
}

function stubApi(status: number, body: unknown) {
  const fetch = vi.fn(async (_request: Request) => Response.json(body, { status }))
  vi.stubGlobal('fetch', fetch)
  return fetch
}

function renderInbox(path = '/inbox', role = 'staff_operative') {
  const user = { id: 'u1', email: 'operativo@fi.uba.ar', name: 'Operativo', role }
  const router = createMemoryRouter(
    [
      {
        element: <Outlet context={user} />,
        children: [
          { path: '/inbox', element: <InboxPage /> },
          { path: '/procedures/:id', element: <ProcedurePage /> },
        ],
      },
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
    renderInbox()

    await tableRows()
    expect(fetch).toHaveBeenCalledTimes(1)
    expect(Object.fromEntries(new URL(fetch.mock.calls[0]![0].url).searchParams)).toEqual({ view: 'all' })
  })

  it('shows the procedure type, pending step, who has it, state and age', async () => {
    stubApi(200, { items: [pending], ...options })
    renderInbox()

    const [row] = await tableRows()
    expect(within(row!).getByRole('link', { name: 'Titular Uno' })).toBeVisible()
    expect(within(row!).getByText('Equivalencia')).toBeVisible()
    expect(within(row!).getByText('Registrar expediente')).toBeVisible()
    expect(within(row!).getByText('Dirección de Alumnos')).toBeVisible()
    expect(within(row!).getByText('En la Secretaría')).toBeVisible()
    expect(within(row!).getByText('3 días')).toBeVisible()
    expect(within(row!).getByText('Hace 3 días')).toBeVisible()
    expect(screen.getByText('Perfil operativo')).toBeVisible()
  })

  it('shows only what needs action by default', async () => {
    stubApi(200, { items: [pending, referred, waiting], ...options })
    renderInbox()

    await tableRows()
    expect(listedNames()).toEqual(['Titular Uno'])
    expect(screen.getByRole('radio', { name: 'Asignados (1)' })).toBeChecked()
    expect(screen.getByRole('radio', { name: 'Todos (3)' })).not.toBeChecked()
  })

  it('counts each state over everything visible, whatever the filters', async () => {
    stubApi(200, { items: [pending, referred, waiting], ...options })
    renderInbox('/inbox?view=all&type=3')

    await tableRows()
    expect(screen.getByRole('button', { name: /Acción requerida\s*1/ })).toBeVisible()
    expect(screen.getByRole('button', { name: /En la Secretaría\s*1/ })).toBeVisible()
    expect(screen.getByRole('button', { name: /En evaluación externa\s*1/ })).toBeVisible()
    expect(screen.getByRole('button', { name: /Pendiente del titular\s*1/ })).toBeVisible()
  })

  it('lists everything in progress, marking what needs action, oldest first', async () => {
    stubApi(200, { items: [pending, referred, waiting], ...options })
    renderInbox('/inbox?view=all')

    const rows = await tableRows()
    expect(listedNames()).toEqual(['Titular Dos', 'Titular Tres', 'Titular Uno'])
    expect(within(rows[2]!).getByText('Acción requerida')).toBeVisible()
    expect(within(rows[0]!).queryByText('Acción requerida')).not.toBeInTheDocument()
    expect(within(rows[0]!).getByText('12 días sin retorno')).toBeVisible()
  })

  it('filters by state from a summary and by the selects, combined', async () => {
    stubApi(200, { items: [pending, referred, waiting], ...options })
    const router = renderInbox()
    await tableRows()

    await userEvent.click(screen.getByRole('button', { name: /En evaluación externa/ }))
    expect(listedNames()).toEqual(['Titular Dos'])

    await userEvent.click(screen.getByRole('combobox', { name: 'Tipo de trámite' }))
    await userEvent.click(screen.getByRole('option', { name: 'Créditos' }))
    expect(listedNames()).toEqual([])
    expect(screen.getByText('No hay trámites que coincidan con los filtros.')).toBeVisible()
    expect(Object.fromEntries(new URLSearchParams(router.state.location.search))).toEqual({
      view: 'all',
      state: 'referred',
      type: '3',
    })
  })

  it('sorts by the last movement', async () => {
    stubApi(200, { items: [pending, referred, waiting], ...options })
    renderInbox('/inbox?view=all&sort=last_moved')

    await tableRows()
    expect(listedNames()).toEqual(['Titular Tres', 'Titular Uno', 'Titular Dos'])
  })

  it('keeps the filters when coming back from a procedure', async () => {
    stubApi(200, { items: [pending, referred, waiting], ...options })
    const router = renderInbox('/inbox?view=all&area=office')

    const [row] = await tableRows()
    await userEvent.click(within(row!).getByRole('link', { name: 'Titular Uno' }))
    expect(router.state.location.pathname).toBe('/procedures/p-1')
    await userEvent.click(screen.getByRole('link', { name: 'Volver a la bandeja de trámites' }))

    await tableRows()
    expect(router.state.location.search).toBe('?view=all&area=office')
    expect(listedNames()).toEqual(['Titular Uno'])
  })

  it('says when nothing needs action', async () => {
    stubApi(200, { items: [referred], ...options })
    renderInbox()

    expect(await screen.findByText('No hay trámites que requieran acción de este perfil.')).toBeVisible()
  })

  it('says when there is nothing in progress', async () => {
    stubApi(200, { items: [], procedure_types: [], areas: [] })
    renderInbox('/inbox?view=all')

    expect(await screen.findByText('No hay trámites en curso.')).toBeVisible()
  })

  it('tells who is not staff that the inbox is not theirs', async () => {
    stubApi(403, { message: 'La bandeja es solo para el personal de la Secretaría.' })
    renderInbox('/inbox', 'applicant')

    expect(await screen.findByRole('alert')).toHaveTextContent('disponible solo para el personal de la Secretaría')
  })

  it('says when the inbox fails', async () => {
    stubApi(500, { message: 'Ocurrió un error inesperado.' })
    renderInbox()

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudo cargar la bandeja de trámites.')
  })

  it('ignores filters in the URL that it does not know', async () => {
    stubApi(200, { items: [pending, referred], ...options })
    renderInbox('/inbox?view=everything&state=approved&type=abc&sort=name')

    await tableRows()
    expect(listedNames()).toEqual(['Titular Uno'])
  })
})
