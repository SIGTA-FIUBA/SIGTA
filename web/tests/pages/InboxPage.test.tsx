import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { InboxPage } from '../../src/pages/InboxPage'
import { ProcedurePage } from '../../src/pages/ProcedurePage'
import { render } from '../render'

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
  started_at: '2026-11-05T12:00:00Z',
  last_moved_at: '2026-11-05T12:00:00Z',
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
  age_days: 30,
  days_referred: 12,
}

const options = {
  procedure_types: [{ id: '2', name: 'Equivalencia' }],
  areas: [
    { id: 'committee', name: 'Comisión Curricular' },
    { id: 'office', name: 'Dirección de Alumnos' },
  ],
}

// stubApi answers /procedures with the items given, and keeps the queries.
function stubApi(status: number, body: unknown) {
  const fetch = vi.fn(async (_request: Request) => Response.json(body, { status }))
  vi.stubGlobal('fetch', fetch)
  return fetch
}

function lastQuery(fetch: ReturnType<typeof stubApi>) {
  const request = fetch.mock.calls.at(-1)![0]
  return Object.fromEntries(new URL(request.url).searchParams)
}

function renderInbox(path = '/inbox') {
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

describe('InboxPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('shows the procedure type, pending step, who has it and its age', async () => {
    stubApi(200, { items: [pending], ...options })
    renderInbox()

    const item = (await screen.findByRole('link', { name: 'Equivalencia: Titular Uno' })).closest('article')!
    expect(
      within(item).getByText('Paso pendiente: Registrar expediente, a cargo de Dirección de Alumnos.'),
    ).toBeVisible()
    expect(within(item).getByText('Empezó hace 3 días.')).toBeVisible()
    expect(within(item).getByText('En la Secretaría')).toBeVisible()
    expect(screen.getByText('1 trámite')).toBeVisible()
  })

  it('shows how long a referred procedure has been without return', async () => {
    stubApi(200, { items: [referred], ...options })
    renderInbox('/inbox?view=all')

    const item = (await screen.findByRole('link', { name: 'Equivalencia: Titular Dos' })).closest('article')!
    expect(within(item).getByText('Empezó hace 30 días. Derivado hace 12 días, sin retorno.')).toBeVisible()
    expect(within(item).getByText('Derivado')).toBeVisible()
  })

  it('asks for what the profile can act on by default, the oldest first', async () => {
    const fetch = stubApi(200, { items: [], ...options })
    renderInbox()

    expect(await screen.findByText('No tenés trámites pendientes.')).toBeVisible()
    expect(lastQuery(fetch)).toEqual({ view: 'mine', sort: 'oldest' })
  })

  it('marks what the profile can act on when listing all', async () => {
    stubApi(200, { items: [pending, referred], ...options })
    renderInbox('/inbox?view=all')

    const first = (await screen.findByRole('link', { name: 'Equivalencia: Titular Uno' })).closest('article')!
    const second = screen.getByRole('link', { name: 'Equivalencia: Titular Dos' }).closest('article')!
    expect(within(first).getByText('Te toca')).toBeVisible()
    expect(within(second).queryByText('Te toca')).not.toBeInTheDocument()
  })

  it('combines the filters and the sort in the query and the URL', async () => {
    const fetch = stubApi(200, { items: [pending], ...options })
    const router = renderInbox()
    await screen.findByText('1 trámite')

    await userEvent.click(screen.getByRole('radio', { name: 'Todos' }))
    await userEvent.click(screen.getByRole('combobox', { name: 'Estado' }))
    await userEvent.click(screen.getByRole('option', { name: 'Derivado' }))
    await userEvent.click(screen.getByRole('combobox', { name: 'Responsable' }))
    await userEvent.click(screen.getByRole('option', { name: 'Comisión Curricular' }))
    await userEvent.click(screen.getByRole('combobox', { name: 'Ordenar por' }))
    await userEvent.click(screen.getByRole('option', { name: 'Movidos más recientemente' }))

    expect(lastQuery(fetch)).toEqual({ view: 'all', state: 'referred', area: 'committee', sort: 'last_moved' })
    expect(Object.fromEntries(new URLSearchParams(router.state.location.search))).toEqual({
      view: 'all',
      state: 'referred',
      area: 'committee',
      sort: 'last_moved',
    })
  })

  it('keeps the filters when coming back from a procedure', async () => {
    const fetch = stubApi(200, { items: [pending], ...options })
    const router = renderInbox('/inbox?view=all&type=2&sort=last_moved')

    await userEvent.click(await screen.findByRole('link', { name: 'Equivalencia: Titular Uno' }))
    expect(router.state.location.pathname).toBe('/procedures/p-1')
    await userEvent.click(screen.getByRole('link', { name: 'Volver a la bandeja' }))

    expect(await screen.findByRole('heading', { name: 'Bandeja' })).toBeVisible()
    expect(router.state.location.search).toBe('?view=all&type=2&sort=last_moved')
    expect(lastQuery(fetch)).toEqual({ view: 'all', procedure_type: '2', sort: 'last_moved' })
  })

  it('says when no procedure matches the filters', async () => {
    stubApi(200, { items: [], ...options })
    renderInbox('/inbox?view=all&state=referred')

    expect(await screen.findByText('No hay trámites con estos filtros.')).toBeVisible()
  })

  it('says when there is nothing in progress', async () => {
    stubApi(200, { items: [], procedure_types: [], areas: [] })
    renderInbox('/inbox?view=all')

    expect(await screen.findByText('No hay trámites en curso.')).toBeVisible()
  })

  it('tells an applicant the inbox is only for the staff', async () => {
    stubApi(403, { message: 'La bandeja es solo para el personal de la Secretaría.' })
    renderInbox()

    expect(await screen.findByRole('alert')).toHaveTextContent('La bandeja es solo para el personal de la Secretaría.')
  })

  it('says when the inbox fails', async () => {
    stubApi(500, { message: 'Ocurrió un error inesperado.' })
    renderInbox()

    expect(await screen.findByRole('alert')).toHaveTextContent('No pudimos cargar la bandeja. Volvé a intentarlo')
  })

  it('ignores filters in the URL that the API does not know', async () => {
    const fetch = stubApi(200, { items: [], ...options })
    renderInbox('/inbox?view=everything&state=approved&type=abc&sort=name')

    await screen.findByText('No tenés trámites pendientes.')
    expect(lastQuery(fetch)).toEqual({ view: 'mine', sort: 'oldest' })
  })
})
