import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ProcedureTypesPage } from '../../src/pages/ProcedureTypesPage'
import { render } from '../test-utils/render'

function draft(changes: Record<string, unknown> = {}) {
  return {
    procedure_type_id: 7,
    base: 0,
    revision: 3,
    status: 'editing',
    definition: { name: 'Pase de universidad extranjera', steps: [] },
    edited_by: { id: '5RnF8i6-BEyFCTX4fs2BN', name: 'Ana Prueba' },
    edited_at: '2026-11-20T12:00:00Z',
    editors: [{ id: '5RnF8i6-BEyFCTX4fs2BN', name: 'Ana Prueba' }],
    problems: [],
    can_edit: true,
    can_review: false,
    ...changes,
  }
}

const published = {
  procedure_type_id: 1,
  name: 'Equivalencia entre asignaturas',
  version: 3,
  published_at: '2026-11-02T12:00:00Z',
  can_open_draft: true,
}

// stubApi answers GET /editable-procedure-types with editable, and every
// change from changes, keyed by method and path.
function stubApi(editable: unknown, changes: Record<string, { status: number; body?: unknown }> = {}) {
  const fetch = vi.fn(async (request: Request) => {
    const path = new URL(request.url).pathname.replace(/^\/api/, '')
    if (request.method === 'GET') {
      return path === '/editable-procedure-types'
        ? Response.json(editable)
        : Response.json({ message: 'No existe.' }, { status: 404 })
    }
    const answer = changes[`${request.method} ${path}`] ?? { status: 204 }
    return answer.body === undefined
      ? new Response(null, { status: answer.status })
      : Response.json(answer.body, { status: answer.status })
  })
  vi.stubGlobal('fetch', fetch)
  return fetch
}

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const router = createMemoryRouter(
    [
      { path: '/procedure-types', element: <ProcedureTypesPage /> },
      { path: '/procedure-types/:procedureTypeId/draft', element: <p>Editor del borrador</p> },
    ],
    { initialEntries: ['/procedure-types'] },
  )
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

describe('ProcedureTypesPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('lists each procedure type with its version, its draft and what the person can do', async () => {
    stubApi({
      can_create: true,
      areas: [],
      procedure_types: [
        published,
        {
          procedure_type_id: 7,
          name: 'Pase de universidad extranjera',
          version: 0,
          can_open_draft: false,
          draft: draft(),
        },
        {
          procedure_type_id: 8,
          name: 'Reconocimiento de créditos',
          version: 1,
          published_at: '2026-11-14T12:00:00Z',
          can_open_draft: false,
          draft: draft({ procedure_type_id: 8, status: 'proposed', can_edit: false, can_review: true }),
        },
      ],
    })
    renderPage()

    const equivalence = (await screen.findByText('Equivalencia entre asignaturas')).closest('li')!
    expect(within(equivalence).getByText('Versión 3, publicada el 2/11/26')).toBeVisible()
    expect(within(equivalence).getByText('Sin borrador')).toBeVisible()
    expect(within(equivalence).getByRole('button', { name: 'Editar versión nueva' })).toBeVisible()
    const foreign = screen.getByText('Pase de universidad extranjera').closest('li')!
    expect(within(foreign).getByText('Sin publicar')).toBeVisible()
    expect(within(foreign).getByText('En edición')).toBeVisible()
    expect(within(foreign).getByRole('link', { name: 'Abrir borrador' })).toHaveAttribute(
      'href',
      '/procedure-types/7/draft',
    )
    const credits = screen.getByText('Reconocimiento de créditos').closest('li')!
    expect(within(credits).getByText('Propuesto')).toBeVisible()
    expect(within(credits).getByText('Pendiente de aprobación')).toBeVisible()
    expect(within(credits).getByRole('link', { name: 'Revisar propuesta' })).toBeVisible()
    expect(screen.getByRole('button', { name: 'Nuevo tipo de trámite' })).toBeVisible()
  })

  it('offers nothing to create or open to whoever only approves', async () => {
    stubApi({ can_create: false, areas: [], procedure_types: [{ ...published, can_open_draft: false }] })
    renderPage()

    expect(await screen.findByText('Equivalencia entre asignaturas')).toBeVisible()
    expect(screen.queryByRole('button', { name: 'Nuevo tipo de trámite' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Editar versión nueva' })).not.toBeInTheDocument()
  })

  it('says when there is nothing to edit or approve', async () => {
    stubApi({ can_create: false, areas: [], procedure_types: [] })
    renderPage()

    expect(await screen.findByText('No hay tipos de trámite que puedas editar o aprobar.')).toBeVisible()
  })

  it('opens the draft of a new version and goes to it', async () => {
    const fetch = stubApi(
      { can_create: false, areas: [], procedure_types: [published] },
      { 'POST /procedure-types/1/draft': { status: 201, body: draft({ procedure_type_id: 1 }) } },
    )
    renderPage()

    await userEvent.click(await screen.findByRole('button', { name: 'Editar versión nueva' }))

    expect(await screen.findByText('Editor del borrador')).toBeVisible()
    expect(fetch.mock.calls.some(([request]) => request.method === 'POST')).toBe(true)
  })

  it('creates a new procedure type with its name and goes to its draft', async () => {
    const fetch = stubApi(
      { can_create: true, areas: [], procedure_types: [] },
      { 'POST /procedure-types': { status: 201, body: draft({ procedure_type_id: 9 }) } },
    )
    renderPage()

    await userEvent.click(await screen.findByRole('button', { name: 'Nuevo tipo de trámite' }))
    await userEvent.type(screen.getByRole('textbox', { name: /Nombre/ }), 'Intercambio')
    await userEvent.click(screen.getByRole('button', { name: 'Crear el borrador' }))

    expect(await screen.findByText('Editor del borrador')).toBeVisible()
    const created = fetch.mock.calls.map(([request]) => request).find((request) => request.method === 'POST')!
    expect(await created.json()).toEqual({ definition: { name: 'Intercambio', steps: [] } })
  })

  it('shows why opening a draft was refused', async () => {
    stubApi(
      { can_create: false, areas: [], procedure_types: [published] },
      {
        'POST /procedure-types/1/draft': {
          status: 409,
          body: { message: 'El tipo de trámite ya tiene un borrador: editá ese.' },
        },
      },
    )
    renderPage()

    await userEvent.click(await screen.findByRole('button', { name: 'Editar versión nueva' }))

    expect(await screen.findByText('El tipo de trámite ya tiene un borrador: editá ese.')).toBeVisible()
  })
})
