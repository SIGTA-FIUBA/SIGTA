import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ProcedureTypeDraftPage } from '../../src/pages/ProcedureTypeDraftPage'
import { render } from '../test-utils/render'

const definition = {
  name: 'Equivalencia entre asignaturas',
  description: 'Reconoce asignaturas aprobadas en otra carrera.',
  fields: [
    { id: 'subjects', name: 'Asignaturas', type: 'subject_list' },
    { id: 'plan', name: 'Plan de estudios', type: 'text' },
  ],
  documents: [
    { id: 'resolution', name: 'Resolución', accepted_types: ['application/pdf'], max_size: 5242880, resolution: true },
  ],
  steps: [
    { id: 'request', name: 'Solicitud', role: 'applicant', action: 'enter_data', fields: ['subjects'], next: 'review' },
    {
      id: 'review',
      name: 'Dictamen de la comisión',
      role: 'staff',
      action: 'decide',
      permissions: [],
      outcomes: [
        { id: 'go', name: 'Continuar', next: 'close', send_back: false },
        { id: 'back', name: 'Pedir información', next: 'request', send_back: true },
      ],
    },
    {
      id: 'close',
      name: 'Resolución',
      role: 'staff',
      area: 'secretariat_office',
      permissions: ['resolve'],
      action: 'upload_document',
      documents: ['resolution'],
    },
  ],
}

// Identifiers are opaque, as those of Google accounts: the page shows names.
const ana = { id: '5RnF8i6-BEyFCTX4fs2BN', name: 'Ana Prueba' }
const beto = { id: '3Las2QqynzRcykV5BlT4P', name: 'Beto Ejemplo' }

const problems = [
  {
    code: 'no_permissions',
    path: [{ kind: 'step', id: 'review' }],
    text: 'staff steps ask for at least one permission',
  },
  { code: 'unused_field', path: [{ kind: 'field', id: 'plan' }], text: 'field "plan" is not used by any step' },
]

function draft(changes: Record<string, unknown> = {}) {
  return {
    procedure_type_id: 1,
    base: 3,
    revision: 7,
    status: 'editing',
    definition,
    edited_by: ana,
    edited_at: '2026-11-20T12:00:00Z',
    editors: [ana],
    problems,
    can_edit: true,
    can_review: false,
    ...changes,
  }
}

const editable = {
  can_create: true,
  procedure_types: [],
  areas: [
    { id: 'commission', name: 'Comisión Curricular', permissions: [{ id: 'evaluate', name: 'Dictaminar' }] },
    { id: 'secretariat_office', name: 'Secretaría', permissions: [{ id: 'resolve', name: 'Resolver' }] },
  ],
}

// stubApi answers the draft and what the person can edit, and every change
// from changes, keyed by method and path.
function stubApi(answer: unknown, changes: Record<string, { status: number; body?: unknown }> = {}) {
  const fetch = vi.fn(async (request: Request) => {
    const path = new URL(request.url).pathname.replace(/^\/api/, '')
    if (request.method === 'GET') {
      return Response.json(path === '/editable-procedure-types' ? editable : answer)
    }
    const change = changes[`${request.method} ${path}`] ?? { status: 204 }
    return change.body === undefined
      ? new Response(null, { status: change.status })
      : Response.json(change.body, { status: change.status })
  })
  vi.stubGlobal('fetch', fetch)
  return fetch
}

async function sent(fetch: ReturnType<typeof stubApi>) {
  const requests = fetch.mock.calls.map(([request]) => request).filter((request) => request.method !== 'GET')
  return Promise.all(
    requests.map(async (request) => ({
      request: `${request.method} ${new URL(request.url).pathname.replace(/^\/api/, '')}`,
      body: request.headers.get('Content-Type') === 'application/json' ? await request.json() : undefined,
    })),
  )
}

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const router = createMemoryRouter(
    [
      { path: '/procedure-types/:procedureTypeId/draft', element: <ProcedureTypeDraftPage /> },
      { path: '/procedure-types', element: <p>Lista de tipos de trámite</p> },
    ],
    { initialEntries: ['/procedure-types/1/draft'] },
  )
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

const step = (name: RegExp) => screen.findByRole('button', { name })
const panel = () => within(screen.getByRole('region', { name: 'Recorrido del tipo de trámite' }))

describe('ProcedureTypeDraftPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('draws the route as a diagram and marks the steps with problems', async () => {
    stubApi(draft())
    renderPage()

    expect(await step(/^Paso 1: Solicitud\. Titular/)).toBeVisible()
    expect(screen.getByRole('button', { name: /^Paso 2: Dictamen de la comisión\. Con problemas/ })).toBeVisible()
    expect(screen.getByRole('button', { name: /^Paso 3: Resolución\. Secretaría/ })).toBeVisible()
    expect(screen.getByText('↩ Pedir información')).toBeVisible()
    expect(screen.getByText('Hay 2 problemas antes de proponer:')).toBeVisible()
    expect(screen.getByRole('button', { name: 'Proponer' })).toBeDisabled()
  })

  it('goes to where a problem is, with its text in Spanish', async () => {
    stubApi(draft())
    renderPage()

    await userEvent.click(
      await screen.findByRole('button', { name: /«Dictamen de la comisión»: Un paso del personal exige/ }),
    )
    expect(panel().getByRole('textbox', { name: 'Nombre' })).toHaveValue('Dictamen de la comisión')
    expect(panel().getByText('Un paso del personal exige al menos un permiso.')).toBeVisible()

    await userEvent.click(screen.getByRole('button', { name: /«Plan de estudios»: Ningún paso pide este dato/ }))
    expect(screen.getByDisplayValue('Plan de estudios')).toBeVisible()
    expect(screen.getAllByText('Ningún paso pide este dato.').length).toBeGreaterThan(0)
  })

  it('edits the selected step and saves the definition on its revision', async () => {
    const fetch = stubApi(draft(), { 'PUT /procedure-types/1/draft': { status: 200, body: draft({ revision: 8 }) } })
    renderPage()

    await userEvent.click(await step(/^Paso 3: Resolución/))
    const name = panel().getByRole('textbox', { name: 'Nombre' })
    await userEvent.clear(name)
    await userEvent.type(name, 'Fin')
    expect(screen.getByText(/Hay cambios sin guardar/)).toBeVisible()
    expect(screen.getByRole('button', { name: 'Proponer' })).toBeDisabled()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    const [save] = await sent(fetch)
    expect(save.request).toBe('PUT /procedure-types/1/draft')
    expect(save.body.revision).toBe(7)
    expect(save.body.definition.steps[2].name).toBe('Fin')
  })

  it('removes a step and leads what led to it where it went', async () => {
    stubApi(draft())
    renderPage()

    await userEvent.click(await step(/^Paso 3: Resolución/))
    await userEvent.click(panel().getByRole('button', { name: 'Quitar este paso' }))
    await userEvent.click(screen.getByRole('button', { name: 'Quitar el paso' }))

    expect(screen.queryByRole('button', { name: /^Paso 3:/ })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeEnabled()
  })

  it('offers whoever can approve to approve or reject a proposal, and nothing to edit', async () => {
    const fetch = stubApi(
      draft({
        status: 'proposed',
        proposed_by: ana,
        proposed_at: '2026-11-21T16:40:00Z',
        can_edit: false,
        can_review: true,
      }),
      { 'POST /procedure-types/1/draft/rejection': { status: 200, body: draft({ revision: 9 }) } },
    )
    renderPage()

    expect(await screen.findByText(/Propuesto por Ana Prueba/)).toBeVisible()
    expect(screen.queryByText(new RegExp(ana.id))).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Aprobar y publicar' })).toBeVisible()
    expect(screen.queryByRole('button', { name: 'Guardar' })).not.toBeInTheDocument()
    expect(panel().getByRole('textbox', { name: 'Nombre' })).toBeDisabled()

    await userEvent.click(screen.getByRole('button', { name: 'Rechazar' }))
    await userEvent.type(screen.getByRole('textbox', { name: /Motivo/ }), 'Falta el plazo de presentación')
    await userEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Rechazar' }))

    const [rejection] = await sent(fetch)
    expect(rejection).toEqual({
      request: 'POST /procedure-types/1/draft/rejection',
      body: { revision: 7, reason: 'Falta el plazo de presentación' },
    })
  })

  it('does not offer to approve to whoever proposed it', async () => {
    stubApi(draft({ status: 'proposed', proposed_by: ana, can_edit: false, can_review: false }))
    renderPage()

    expect(await screen.findByText(/No se edita mientras espera/)).toBeVisible()
    expect(screen.queryByRole('button', { name: 'Aprobar y publicar' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Proponer' })).not.toBeInTheDocument()
  })

  it('shows who rejected the draft and why', async () => {
    stubApi(draft({ rejection: { by: beto, at: '2026-11-21T17:05:00Z', reason: 'Falta el plazo' } }))
    renderPage()

    expect(await screen.findByText(/Rechazado por Beto Ejemplo/)).toBeVisible()
    expect(screen.getByText('Motivo: «Falta el plazo»')).toBeVisible()
  })

  it('publishes an approved proposal and goes back to the list', async () => {
    const fetch = stubApi(draft({ status: 'proposed', problems: [], can_edit: false, can_review: true }), {
      'POST /procedure-types/1/draft/approval': {
        status: 200,
        body: {
          procedure_type_id: 1,
          number: 4,
          digest: 'x',
          proposed_by: 'ana',
          approved_by: 'beto',
          published_at: '2026-11-21T17:20:00Z',
        },
      },
    })
    renderPage()

    await userEvent.click(await screen.findByRole('button', { name: 'Aprobar y publicar' }))
    await userEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Aprobar y publicar' }))

    expect(await screen.findByText('Lista de tipos de trámite')).toBeVisible()
    expect((await sent(fetch))[0]).toEqual({ request: 'POST /procedure-types/1/draft/approval', body: { revision: 7 } })
  })

  it('keeps the changes when another person saves meanwhile, and says so', async () => {
    const fetch = stubApi(draft(), {
      'PUT /procedure-types/1/draft': {
        status: 409,
        body: { message: 'El borrador cambió desde que lo abriste. Volvé a cargarlo.' },
      },
    })
    renderPage()

    await userEvent.click(await step(/^Paso 3: Resolución/))
    const name = panel().getByRole('textbox', { name: 'Nombre' })
    await userEvent.clear(name)
    await userEvent.type(name, 'Fin')
    // Someone else saves: the draft read again comes at a later revision.
    fetch.mockImplementation(async (request: Request) => {
      const path = new URL(request.url).pathname.replace(/^\/api/, '')
      if (request.method === 'GET') {
        return Response.json(path === '/editable-procedure-types' ? editable : draft({ revision: 8 }))
      }
      return Response.json({ message: 'El borrador cambió desde que lo abriste. Volvé a cargarlo.' }, { status: 409 })
    })
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByText('Otra persona guardó este borrador mientras lo editabas')).toBeVisible()
    expect(panel().getByRole('textbox', { name: 'Nombre' })).toHaveValue('Fin')
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeDisabled()

    await userEvent.click(screen.getByRole('button', { name: 'Descartar mis cambios y cargar la última versión' }))
    expect(panel().getByRole('textbox', { name: 'Nombre' })).toHaveValue('Resolución')
  })

  it('takes the revision it saved as its own and keeps what changes while saving', async () => {
    const fetch = stubApi(draft())
    renderPage()
    await userEvent.click(await step(/^Paso 3: Resolución/))
    const name = panel().getByRole('textbox', { name: 'Nombre' })
    await userEvent.clear(name)
    await userEvent.type(name, 'Fin')
    // The save and, after it, what the person can edit wait to answer: the
    // draft is read again at the saved revision in between.
    let latest = draft()
    let answerSave = () => {}
    let answerEditable = () => {}
    const waiting = (set: (answer: () => void) => void) => new Promise<void>((resolve) => set(resolve))
    fetch.mockImplementation(async (request: Request) => {
      const path = new URL(request.url).pathname.replace(/^\/api/, '')
      if (request.method === 'PUT') {
        const body = await request.clone().json()
        await waiting((answer) => (answerSave = answer))
        latest = draft({ revision: body.revision + 1, definition: body.definition })
        return Response.json(latest)
      }
      if (path === '/editable-procedure-types') {
        await waiting((answer) => (answerEditable = answer))
        return Response.json(editable)
      }
      return Response.json(latest)
    })
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    await userEvent.type(name, 'al')
    answerSave()

    const reads = () => fetch.mock.calls.filter(([request]) => request.url.endsWith('/procedure-types/1/draft'))
    await vi.waitFor(() => expect(reads().length).toBeGreaterThan(2))
    expect(screen.queryByText('Otra persona guardó este borrador mientras lo editabas')).not.toBeInTheDocument()
    answerEditable()
    expect(name).toHaveValue('Final')
    expect(await screen.findByText(/Hay cambios sin guardar/)).toBeVisible()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    const [, again] = await sent(fetch)
    expect(again.body.revision).toBe(8)
    expect(again.body.definition.steps[2].name).toBe('Final')
  })

  it('shows what it saved while the draft is read again, and proposes on the saved revision', async () => {
    const fetch = stubApi(draft({ problems: [] }))
    renderPage()
    await userEvent.click(await step(/^Paso 3: Resolución/))
    const name = panel().getByRole('textbox', { name: 'Nombre' })
    await userEvent.clear(name)
    await userEvent.type(name, 'Fin')
    // The draft read again after saving takes long to come.
    let latest = draft({ problems: [] })
    let answerRead = () => {}
    fetch.mockImplementation(async (request: Request) => {
      const path = new URL(request.url).pathname.replace(/^\/api/, '')
      if (request.method === 'PUT') {
        const body = await request.clone().json()
        latest = draft({ revision: body.revision + 1, definition: body.definition, problems: [] })
        return Response.json(latest)
      }
      if (request.method === 'POST') {
        return new Response(null, { status: 204 })
      }
      if (path === '/editable-procedure-types') {
        return Response.json(editable)
      }
      await new Promise<void>((resolve) => (answerRead = resolve))
      return Response.json(latest)
    })
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    await vi.waitFor(() => expect(screen.queryByText(/Hay cambios sin guardar/)).not.toBeInTheDocument())
    expect(name).toHaveValue('Fin')
    expect(screen.queryByText('Otra persona guardó este borrador mientras lo editabas')).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Proponer' }))
    await userEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Proponer' }))

    await vi.waitFor(async () =>
      expect((await sent(fetch)).at(-1)).toEqual({
        request: 'POST /procedure-types/1/draft/proposal',
        body: { revision: 8 },
      }),
    )
    answerRead()
  })

  it('shows the reviewer the problems a proposal has with the areas of today', async () => {
    stubApi(draft({ status: 'proposed', proposed_by: ana, can_edit: false, can_review: true }))
    renderPage()

    expect(await screen.findByText(/La propuesta tiene 2 problemas con las áreas de hoy/)).toBeVisible()
  })

  it('offers an outcome only another step to lead to, never closing the procedure', async () => {
    stubApi(draft())
    renderPage()

    await userEvent.click(await step(/^Paso 2: Dictamen/))
    await userEvent.click(panel().getAllByLabelText('Lleva a')[0])

    expect(screen.getByRole('option', { name: 'Solicitud' })).toBeVisible()
    expect(screen.queryByRole('option', { name: 'Cierra el trámite' })).not.toBeInTheDocument()
  })
})
