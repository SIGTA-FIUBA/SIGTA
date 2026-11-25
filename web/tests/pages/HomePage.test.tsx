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

const mine = (id: string, name: string, moved: string, extra: object = {}) => ({
  id,
  procedure_type: 1,
  procedure_type_name: name,
  step_name: 'Revisión de la solicitud',
  state: 'in_office',
  correction_pending: false,
  started_at: '2026-11-01T12:00:00Z',
  last_moved_at: moved,
  resolutions: [],
  ...extra,
})

const procedureType = (id: number) => ({
  id,
  version: 1,
  name: `Trámite ${id}`,
  description: 'Una descripción',
  deadlines: [],
  admissibility: [],
  fields: [{ id: 'name', name: 'Nombre', required: true }],
  documents: [],
  steps: [],
})

// The detail of a procedure with a correction that asks for the subjects.
const awaiting = {
  id: 'P-2',
  procedure_type: 1,
  version: 1,
  name: 'Reconocimiento de créditos',
  state: 'with_applicant',
  applicant: { id: 'u1', name: 'Persona' },
  open_step: 'review',
  can_act: false,
  as_applicant: true,
  correction_requests: [
    { step: 'review', at: '2026-11-17T15:00:00Z', reason: 'Falta', fields: ['subjects'], documents: [] },
  ],
  steps: [
    {
      id: 'data',
      name: 'Datos',
      action: 'enter_data',
      role: 'applicant',
      fields: [{ id: 'subjects', name: 'Materias aprobadas', required: true }],
      documents: [],
    },
  ],
}

// stubApi plays the API: /me/procedures answers the procedures the person
// started, /procedures the inbox of the staff, /procedure-types the catalog
// and /procedures/P-2 the detail of a procedure with a correction.
function stubApi({
  procedures = [],
  inbox = [],
  catalog = [],
  status = 200,
}: {
  procedures?: unknown[]
  inbox?: unknown[]
  catalog?: unknown[]
  status?: number
}) {
  const fetch = vi.fn(async (request: Request) => {
    if (status !== 200) {
      return Response.json({ message: 'error' }, { status })
    }
    const path = new URL(request.url).pathname
    if (path === '/api/me/procedures') {
      return Response.json(procedures)
    }
    if (path === '/api/procedure-types') {
      return Response.json(catalog)
    }
    if (path === '/api/procedures/P-2') {
      return Response.json(awaiting)
    }
    return Response.json({ items: inbox })
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

  it('titles the page without a greeting', async () => {
    stubApi({})
    renderAs(false)

    expect(screen.getByRole('heading', { level: 1, name: 'Mis trámites' })).toBeInTheDocument()
    expect(screen.queryByText(/Hola/)).not.toBeInTheDocument()
    expect(await screen.findByText('Sin pendientes')).toBeVisible()
  })

  it('lists what each correction asks of the applicant, under the procedure', async () => {
    stubApi({
      procedures: [
        mine('P-1', 'Equivalencia entre asignaturas', '2026-11-10T15:00:00Z'),
        mine('P-2', 'Reconocimiento de créditos', '2026-11-12T15:00:00Z', {
          state: 'with_applicant',
          correction_pending: true,
        }),
      ],
    })
    renderAs(false)

    const pending = await screen.findByRole('list', { name: 'Pendientes' })
    const row = await within(pending).findByRole('link', { name: /Materias aprobadas/ })
    expect(row).toHaveAttribute('href', '/procedures/P-2')
    expect(within(row).getByText('Reconocimiento de créditos')).toBeVisible()
    expect(within(pending).getAllByRole('link')).toHaveLength(1)
  })

  it('lists a procedure waiting on the applicant for a step of theirs, by that step', async () => {
    stubApi({
      procedures: [
        mine('P-9', 'Pase de carrera', '2026-11-12T15:00:00Z', { state: 'with_applicant', step_name: 'Aceptar' }),
      ],
    })
    renderAs(false)

    const pending = await screen.findByRole('list', { name: 'Pendientes' })
    const row = await within(pending).findByRole('link', { name: /Aceptar/ })
    expect(row).toHaveAttribute('href', '/procedures/P-9')
    expect(within(row).getByText('Pase de carrera')).toBeVisible()
  })

  it('shows every procedure in progress, the latest first, and leads to the history', async () => {
    stubApi({
      procedures: [
        mine('P-1', 'Primero', '2026-11-01T15:00:00Z'),
        mine('P-2b', 'Segundo', '2026-11-02T15:00:00Z'),
        mine('P-3', 'Tercero', '2026-11-03T15:00:00Z'),
        mine('P-4', 'Cuarto', '2026-11-04T15:00:00Z'),
        mine('P-5', 'Cerrado', '2026-11-05T15:00:00Z', { state: 'closed' }),
      ],
    })
    renderAs(false)

    const inProgress = await screen.findByRole('list', { name: 'Trámites en curso' })
    const links = within(inProgress).getAllByRole('link')
    expect(links.map((link) => link.getAttribute('href'))).toEqual([
      '/procedures/P-4',
      '/procedures/P-3',
      '/procedures/P-2b',
      '/procedures/P-1',
    ])
    expect(screen.getByRole('link', { name: 'Historial' })).toHaveAttribute('href', '/my-procedures')
  })

  it('offers three procedures of the catalog to start, without what they ask for, and all of them', async () => {
    stubApi({ catalog: Array.from({ length: 8 }, (_, index) => procedureType(index + 1)) })
    renderAs(false)

    const first = await screen.findByRole('link', { name: /Trámite 1/ })
    expect(first).toHaveAttribute('href', '/catalog/1/start')
    expect(within(first).queryByText(/dato/)).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Trámite 3/ })).toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /Trámite 4/ })).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Ver todos' })).toHaveAttribute('href', '/new')
  })

  it('says when the procedures could not be loaded', async () => {
    stubApi({ status: 500 })
    renderAs(false)

    const [message] = await screen.findAllByText(/No se pudieron cargar los trámites/)
    expect(message.closest('[role="alert"]')).not.toBeNull()
  })

  it('puts on top the procedures to start, then what is pending and then those in progress', async () => {
    stubApi({ catalog: [procedureType(1)] })
    renderAs(false)

    await screen.findByRole('link', { name: /Trámite 1/ })
    expect(screen.getAllByRole('heading', { level: 2 }).map((heading) => heading.textContent)).toEqual([
      'Nuevo trámite',
      'Pendientes',
      'Trámites en curso',
    ])
  })

  it('lists for the staff the step to resolve of the procedures that need action, the latest first', async () => {
    stubApi({
      inbox: [
        { ...item, id: 'old' },
        { ...item, id: 'new', applicant: { id: 'a2', name: 'Titular Dos' }, last_moved_at: '2026-11-05T10:00:00Z' },
        { ...item, id: 'waiting', applicant: { id: 'a3', name: 'Titular Tres' }, actionable: false },
      ],
    })
    renderAs(true)

    const list = await screen.findByRole('list', { name: 'Pendientes' })
    const links = within(list).getAllByRole('link')
    expect(links.map((link) => link.getAttribute('href'))).toEqual(['/procedures/new', '/procedures/old'])
    expect(within(links[0]!).getByText('Registrar expediente')).toBeInTheDocument()
    expect(within(links[0]!).getByText('Equivalencia, Titular Dos')).toBeInTheDocument()
    expect(screen.queryByRole('link', { name: 'Ver todos' })).not.toBeInTheDocument()
  })

  it('shows at most five pending procedures and leads to the inbox for the rest', async () => {
    stubApi({ inbox: Array.from({ length: 7 }, (_, index) => ({ ...item, id: `p${index}` })) })
    renderAs(true)

    const list = await screen.findByRole('list', { name: 'Pendientes' })
    expect(within(list).getAllByRole('link')).toHaveLength(5)
    expect(screen.getByRole('link', { name: 'Ver todos' })).toHaveAttribute('href', '/inbox')
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
    expect(lines[0]).toHaveTextContent('Equivalencia de Titular Dos está en Registrar expediente')
    expect(lines[1]).toHaveTextContent('Equivalencia de Titular Uno está en Registrar expediente')
  })

  it('says so when no procedure needs action', async () => {
    stubApi({ inbox: [{ ...item, id: 'waiting', actionable: false }] })
    renderAs(true)

    expect(await screen.findByText('Sin pendientes')).toBeVisible()
  })

  it('says when the inbox of the staff could not be loaded', async () => {
    stubApi({ status: 500 })
    renderAs(true)

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudieron cargar los trámites')
  })
})
