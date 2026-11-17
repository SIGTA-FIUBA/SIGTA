import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { components } from '../../src/api/schema'
import { ProcedurePage } from '../../src/pages/ProcedurePage'
import { render } from '../render'

type Detail = components['schemas']['ProcedureDetail']

const detail: Detail = {
  id: 'P-1',
  procedure_type: 1,
  version: 1,
  name: 'Equivalencia de asignaturas',
  state: 'in_office',
  applicant: { id: 'a-1', name: 'Titular Uno' },
  open_step: 'review',
  can_act: true,
  steps: [
    {
      id: 'documents',
      name: 'Adjuntar documentos',
      action: 'upload_document',
      role: 'applicant',
      fields: [],
      documents: [
        { id: 'statement', name: 'Constancia', required: true, accepted_types: ['application/pdf'], max_size: 1024 },
        { id: 'annex', name: 'Anexo', required: false, accepted_types: ['application/pdf'], max_size: 1024 },
      ],
      outcomes: [],
      next: 'review',
      values: [],
      files: [
        {
          id: 'D-1',
          document: 'statement',
          step: 'documents',
          file_name: 'constancia.pdf',
          media_type: 'application/pdf',
          size: 2048,
          uploaded_by: { id: 'a-1', name: 'Titular Uno' },
          uploaded_at: '2026-11-16T12:00:00Z',
          current: true,
        },
      ],
    },
    {
      id: 'review',
      name: 'Revisar la solicitud',
      action: 'decide',
      role: 'staff',
      area: { id: 'office', name: 'Dirección de Alumnos' },
      fields: [],
      documents: [],
      outcomes: [{ id: 'accept', name: 'Aceptar', next: 'documents' }],
      values: [],
      files: [],
    },
  ],
  history: [],
}

function stubApi(status: number, body: unknown) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async (_request: Request) => Response.json(body, { status })),
  )
}

function renderPage(state?: unknown) {
  const router = createMemoryRouter([{ path: '/procedures/:id', element: <ProcedurePage /> }], {
    initialEntries: [{ pathname: '/procedures/P-1', state }],
  })
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

describe('ProcedurePage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('shows the state, the applicant and the open step', async () => {
    stubApi(200, detail)
    renderPage()

    expect(await screen.findByRole('heading', { level: 1, name: 'Equivalencia de asignaturas' })).toBeDefined()
    expect(screen.getByText('En un área')).toBeDefined()
    expect(screen.getByText('Titular Uno')).toBeDefined()
    expect(screen.getByText('Revisar la solicitud · Dirección de Alumnos')).toBeDefined()
    expect(screen.getByText('Acción requerida')).toBeDefined()
  })

  it('does not ask for action from who cannot act', async () => {
    stubApi(200, { ...detail, can_act: false })
    renderPage()

    await screen.findByRole('heading', { level: 1 })
    expect(screen.queryByText('Acción requerida')).toBeNull()
  })

  it('shows the file of each document next to it', async () => {
    stubApi(200, detail)
    renderPage()

    const link = await screen.findByRole('link', { name: 'constancia.pdf' })
    expect(link.getAttribute('href')).toBe('/api/procedures/P-1/documents/D-1/content')
    expect(screen.getByText(/Anexo:/).textContent).toBe('Anexo: Sin archivo')
  })

  it('shows the data entered once each, as the applicant chose it', async () => {
    const enteredBy = { id: 'a-1', name: 'Titular Uno' }
    const subjects = {
      field: 'subjects',
      step: 'documents',
      items: ['Análisis II', 'Física I'],
      entered_by: enteredBy,
      entered_at: '2026-11-16T12:00:00Z',
    }
    const subjectList = {
      id: 'subjects',
      name: 'Asignaturas',
      required: true,
      type: 'subject_list' as const,
      options: [],
    }
    stubApi(200, {
      ...detail,
      steps: [
        {
          ...detail.steps[0]!,
          fields: [
            { id: 'reason', name: 'Motivo', required: true, type: 'long_text', options: [] },
            {
              id: 'plan',
              name: 'Plan',
              required: true,
              type: 'option',
              options: [{ id: 'p2023', name: 'Plan 2023' }],
            },
            subjectList,
            { id: 'phone', name: 'Teléfono', required: false, type: 'text', options: [] },
          ],
          values: [
            {
              field: 'reason',
              step: 'documents',
              text: 'Pase de carrera\nDesde Sistemas',
              entered_by: enteredBy,
              entered_at: '2026-11-16T12:00:00Z',
            },
            {
              field: 'plan',
              step: 'documents',
              text: 'p2023',
              entered_by: enteredBy,
              entered_at: '2026-11-16T12:00:00Z',
            },
            subjects,
          ],
        },
        { ...detail.steps[1]!, fields: [subjectList], values: [subjects] },
      ],
    })
    renderPage()

    await screen.findByRole('heading', { level: 2, name: 'Datos del trámite' })
    expect(screen.getByText('Pase de carrera')).toBeDefined()
    expect(screen.getByText('Desde Sistemas')).toBeDefined()
    expect(screen.getByText('Plan 2023')).toBeDefined()
    expect(screen.getAllByText('Análisis II, Física I')).toHaveLength(1)
    expect(screen.queryByText('Teléfono')).toBeNull()
  })

  it('says when no data was entered', async () => {
    stubApi(200, detail)
    renderPage()

    expect(await screen.findByText('El trámite no tiene datos cargados.')).toBeDefined()
  })

  it('marks the open step and says who does each step', async () => {
    stubApi(200, detail)
    renderPage()

    await screen.findByRole('heading', { level: 2, name: 'Pasos' })
    const review = screen.getAllByText('Revisar la solicitud').at(-1)!.closest('div.mantine-Paper-root') as HTMLElement
    expect(within(review).getByText('Paso actual')).toBeDefined()
    expect(within(review).getByText('Dirección de Alumnos')).toBeDefined()
  })

  it('says where a referred procedure is and for how long', async () => {
    stubApi(200, {
      ...detail,
      state: 'referred',
      referral: {
        area: { id: 'committee', name: 'Comisión Curricular' },
        from: { id: 'office', name: 'Dirección de Alumnos' },
        days_without_return: 4,
      },
    })
    renderPage()

    expect(await screen.findByText('Derivado')).toBeDefined()
    expect(screen.getByText('En Comisión Curricular, desde Dirección de Alumnos. 4 días sin retorno.')).toBeDefined()
  })

  it('says when the procedure is closed', async () => {
    stubApi(200, { ...detail, state: 'closed', open_step: undefined, can_act: false })
    renderPage()

    expect(await screen.findByText('Cerrado')).toBeDefined()
    expect(screen.getByText('Sin paso abierto')).toBeDefined()
  })

  it('shows the history in the order it was recorded, with who, when and what', async () => {
    stubApi(200, {
      ...detail,
      steps: [
        ...detail.steps,
        {
          ...detail.steps[1]!,
          id: 'evaluate',
          name: 'Dictamen',
          action: 'confirm_external_fact',
          area: { id: 'committee', name: 'Comisión Curricular' },
          outcomes: [],
        },
      ],
      history: [
        {
          step: 'documents',
          by: { id: 'a-1', name: 'Titular Uno' },
          at: '2026-11-16T15:00:00Z',
          observation: '',
          declared: false,
          returned: false,
        },
        {
          step: 'review',
          by: { id: 's-1', name: 'Operativa Dos' },
          at: '2026-11-17T15:30:00Z',
          observation: 'Documentación completa',
          outcome: 'accept',
          declared: false,
          returned: false,
          referral: { from: 'office', to: 'committee', external_file_number: 'EX-2026-00001' },
        },
        {
          step: 'evaluate',
          by: { id: 's-2', name: '' },
          at: '2026-11-20T13:00:00Z',
          observation: 'Dictamen recibido',
          declared: true,
          returned: true,
        },
      ],
    })
    renderPage()

    await screen.findByRole('heading', { level: 2, name: 'Historial' })
    const items = screen.getAllByText(/ · \d{2}\/\d{2}\/\d{4}/).map((item) => item.textContent)
    expect(items).toEqual([
      'Titular Uno · 16/11/2026, 12:00',
      'Operativa Dos · 17/11/2026, 12:30',
      's-2 · 20/11/2026, 10:00',
    ])
    expect(screen.getByText('Salida: Aceptar')).toBeDefined()
    expect(screen.getByText('Documentación completa')).toBeDefined()
    expect(screen.getByText('Derivado a Comisión Curricular, expediente EX-2026-00001')).toBeDefined()
    expect(screen.getByText('Volvió al área de la que se derivó')).toBeDefined()
    expect(screen.getByText('Hecho fuera de SIGTA, declarado por quien lo confirmó')).toBeDefined()
  })

  it('says when no step has been done yet', async () => {
    stubApi(200, detail)
    renderPage()

    expect(await screen.findByText('Todavía no hay pasos hechos.')).toBeDefined()
  })

  it('says when the procedure does not exist or is not theirs', async () => {
    stubApi(404, { message: 'No existe el trámite.' })
    renderPage()

    expect(await screen.findByText(/No existe el trámite/)).toBeDefined()
  })

  it('says when the procedure cannot be loaded', async () => {
    stubApi(500, { message: 'error' })
    renderPage()

    expect((await screen.findByRole('alert')).textContent).toBe(
      'No pudimos cargar el trámite. Volvé a intentarlo más tarde.',
    )
  })

  it('leads back home when it was not opened from the inbox', async () => {
    stubApi(200, detail)
    renderPage()

    expect((await screen.findByRole('link', { name: 'Volver al inicio' })).getAttribute('href')).toBe('/')
  })
})

describe('ProcedurePage confirmation', () => {
  const referable: Detail = {
    ...detail,
    steps: [
      detail.steps[0]!,
      {
        ...detail.steps[1]!,
        outcomes: [
          { id: 'accept', name: 'Aceptar', next: 'evaluate' },
          { id: 'send_back', name: 'Devolver al titular', next: 'documents' },
        ],
      },
      {
        ...detail.steps[1]!,
        id: 'evaluate',
        name: 'Dictamen',
        action: 'confirm_external_fact',
        area: { id: 'committee', name: 'Comisión Curricular' },
        outcomes: [],
      },
    ],
  }

  // stubConfirmation answers the detail, and the confirmation with status.
  function stubConfirmation(status: number, body: unknown, current: Detail = referable) {
    const fetch = vi.fn(async (request: Request) =>
      request.method === 'POST' ? Response.json(body, { status }) : Response.json(current),
    )
    vi.stubGlobal('fetch', fetch)
    return fetch
  }

  async function sent(fetch: ReturnType<typeof stubConfirmation>) {
    const posts = fetch.mock.calls.map(([request]) => request).filter((request) => request.method === 'POST')
    return Promise.all(posts.map(async (request) => ({ url: request.url, body: await request.json() })))
  }

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('offers to confirm only an open step of the staff that the person can do', async () => {
    stubConfirmation(201, {})
    renderPage()
    expect(await screen.findByRole('heading', { level: 2, name: 'Confirmar Revisar la solicitud' })).toBeDefined()
  })

  it.each([
    ['who cannot act', { ...referable, can_act: false }],
    ['a step of the applicant', { ...referable, open_step: 'documents', state: 'with_applicant' as const }],
  ])('does not offer to confirm to %s', async (_, current) => {
    stubConfirmation(201, {}, current)
    renderPage()
    await screen.findByRole('heading', { level: 1 })
    expect(screen.queryByRole('button', { name: 'Confirmar el paso' })).toBeNull()
  })

  it('asks for an outcome and an observation before sending', async () => {
    const fetch = stubConfirmation(201, {})
    renderPage()

    await userEvent.click(await screen.findByRole('button', { name: 'Confirmar el paso' }))

    expect(screen.getByText('Elegí una salida.')).toBeDefined()
    expect(screen.getByText('Escribí una observación.')).toBeDefined()
    expect(await sent(fetch)).toEqual([])
  })

  it('asks for the external file number only when the outcome refers the procedure', async () => {
    stubConfirmation(201, {})
    renderPage()

    await userEvent.click(await screen.findByRole('combobox', { name: /Salida/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Devolver al titular' }))
    expect(screen.queryByLabelText('Número de expediente externo')).toBeNull()

    await userEvent.click(screen.getByRole('combobox', { name: /Salida/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Aceptar' }))
    expect(screen.getByLabelText('Número de expediente externo')).toBeDefined()
  })

  it('sends the confirmation and says it is in the history', async () => {
    const fetch = stubConfirmation(201, {})
    renderPage()

    await userEvent.click(await screen.findByRole('combobox', { name: /Salida/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Aceptar' }))
    await userEvent.type(screen.getByRole('textbox', { name: /Observación/ }), ' Documentación completa ')
    await userEvent.type(screen.getByLabelText('Número de expediente externo'), 'EX-2026-00001')
    await userEvent.click(screen.getByRole('button', { name: 'Confirmar el paso' }))

    expect(await screen.findByText('Paso confirmado. Quedó en el historial.')).toBeDefined()
    expect(await sent(fetch)).toEqual([
      {
        url: expect.stringMatching(/\/api\/procedures\/P-1\/steps\/review\/confirmation$/),
        body: { observation: 'Documentación completa', outcome: 'accept', external_file_number: 'EX-2026-00001' },
      },
    ])
    const reads = fetch.mock.calls.filter(([request]) => request.method === 'GET')
    expect(reads.length).toBeGreaterThan(1)
  })

  it('shows why the server did not take the confirmation', async () => {
    stubConfirmation(409, { message: 'El paso ya no está abierto: el trámite cambió mientras tanto.' })
    renderPage()

    await userEvent.click(await screen.findByRole('combobox', { name: /Salida/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Devolver al titular' }))
    await userEvent.type(screen.getByRole('textbox', { name: /Observación/ }), 'Falta la firma')
    await userEvent.click(screen.getByRole('button', { name: 'Confirmar el paso' }))

    expect((await screen.findByRole('alert')).textContent).toBe(
      'El paso ya no está abierto: el trámite cambió mientras tanto.',
    )
    expect(screen.queryByText('Paso confirmado. Quedó en el historial.')).toBeNull()
  })
})
