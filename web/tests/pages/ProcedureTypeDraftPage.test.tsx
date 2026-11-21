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
    edited_by: 'ana@fi.uba.ar',
    edited_at: '2026-11-20T12:00:00Z',
    editors: ['ana@fi.uba.ar'],
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
        proposed_by: 'ana@fi.uba.ar',
        proposed_at: '2026-11-21T16:40:00Z',
        can_edit: false,
        can_review: true,
      }),
      { 'POST /procedure-types/1/draft/rejection': { status: 200, body: draft({ revision: 9 }) } },
    )
    renderPage()

    expect(await screen.findByText(/Propuesto por ana@fi\.uba\.ar/)).toBeVisible()
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
    stubApi(draft({ status: 'proposed', proposed_by: 'ana@fi.uba.ar', can_edit: true, can_review: false }))
    renderPage()

    expect(await screen.findByText(/No se edita mientras espera/)).toBeVisible()
    expect(screen.queryByRole('button', { name: 'Aprobar y publicar' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Proponer' })).not.toBeInTheDocument()
  })

  it('shows who rejected the draft and why', async () => {
    stubApi(draft({ rejection: { by: 'beto@fi.uba.ar', at: '2026-11-21T17:05:00Z', reason: 'Falta el plazo' } }))
    renderPage()

    expect(await screen.findByText(/Rechazado por beto@fi\.uba\.ar/)).toBeVisible()
    expect(screen.getByText(/«Falta el plazo»/)).toBeVisible()
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
})
