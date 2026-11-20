import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { components } from '../../src/api/schema'
import { ProcedurePage } from '../../src/pages/ProcedurePage'
import { render } from '../test-utils/render'

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
  as_applicant: false,
  correction_requests: [],
  steps: [
    {
      id: 'documents',
      name: 'Adjuntar documentos',
      action: 'upload_document',
      role: 'applicant',
      fields: [],
      documents: [
        {
          id: 'statement',
          name: 'Constancia',
          required: true,
          accepted_types: ['application/pdf'],
          max_size: 1024,
          resolution: false,
        },
        {
          id: 'annex',
          name: 'Anexo',
          required: false,
          accepted_types: ['application/pdf'],
          max_size: 1024,
          resolution: false,
        },
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
      outcomes: [{ id: 'accept', name: 'Aceptar', next: 'documents', send_back: false }],
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

// stubPost answers the detail with current, and any POST with status.
function stubPost(status: number, body: unknown, current: Detail) {
  const fetch = vi.fn(async (request: Request) =>
    request.method === 'POST' ? Response.json(body, { status }) : Response.json(current),
  )
  vi.stubGlobal('fetch', fetch)
  return fetch
}

async function sentPosts(fetch: ReturnType<typeof stubPost>) {
  const posts = fetch.mock.calls.map(([request]) => request).filter((request) => request.method === 'POST')
  return Promise.all(posts.map(async (request) => ({ url: request.url, body: await request.json() })))
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

    const link = await screen.findByRole('link', { name: 'Descargar constancia.pdf' })
    expect(link.getAttribute('href')).toBe('/api/procedures/P-1/documents/D-1/content')
    expect(screen.getByText(/^2\sKB$/)).toBeDefined()
    expect(screen.getByText('Anexo').nextElementSibling?.textContent).toBe('Sin archivo')
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
    expect(screen.getAllByText('Análisis II')).toHaveLength(1)
    expect(screen.getAllByText('Física I')).toHaveLength(1)
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

  it('tells the applicant what the procedure waits for now', async () => {
    stubApi(200, { ...detail, as_applicant: true, can_act: false })
    renderPage()

    expect(await screen.findByText('Qué se espera ahora')).toBeDefined()
    expect(screen.getByText('Revisar la solicitud, a cargo de Dirección de Alumnos.')).toBeDefined()
  })

  it('tells the applicant when the procedure waits for them', async () => {
    stubApi(200, { ...detail, as_applicant: true, open_step: 'documents', state: 'with_applicant' })
    renderPage()

    expect(await screen.findByText('Adjuntar documentos, a cargo del titular.')).toBeDefined()
  })

  it('tells the applicant when the procedure is out for evaluation', async () => {
    stubApi(200, {
      ...detail,
      as_applicant: true,
      can_act: false,
      state: 'referred',
      referral: {
        area: { id: 'committee', name: 'Comisión Curricular' },
        from: { id: 'office', name: 'Dirección de Alumnos' },
        days_without_return: 4,
      },
    })
    renderPage()

    expect(
      await screen.findByText(
        'Revisar la solicitud, en evaluación de Comisión Curricular, fuera de Dirección de Alumnos.',
      ),
    ).toBeDefined()
  })

  it('does not tell staff what the procedure waits for, since they see the open step', async () => {
    stubApi(200, detail)
    renderPage()

    await screen.findByText('Revisar la solicitud · Dirección de Alumnos')
    expect(screen.queryByText('Qué se espera ahora')).toBeNull()
  })

  it('offers the resolution once the procedure closes', async () => {
    const resolve = {
      ...detail.steps[1]!,
      id: 'resolve',
      name: 'Resolución',
      action: 'upload_document' as const,
      outcomes: [],
      documents: [
        {
          id: 'resolution',
          name: 'Resolución de Decano',
          required: true,
          accepted_types: ['application/pdf'],
          max_size: 1024,
          resolution: true,
        },
      ],
      files: [
        {
          id: 'D-9',
          document: 'resolution',
          step: 'resolve',
          file_name: 'resolucion.pdf',
          media_type: 'application/pdf',
          size: 2048,
          uploaded_by: { id: 's-1', name: 'Operativa' },
          uploaded_at: '2026-11-20T12:00:00Z',
          current: true,
        },
      ],
    }
    stubApi(200, {
      ...detail,
      as_applicant: true,
      state: 'closed',
      open_step: undefined,
      can_act: false,
      steps: [...detail.steps, resolve],
    })
    renderPage()

    const heading = await screen.findByRole('heading', { level: 2, name: 'Resolución' })
    const section = heading.parentElement!
    expect(within(section).getByText('Resolución de Decano')).toBeDefined()
    expect(within(section).getByText('resolucion.pdf')).toBeDefined()
    expect(within(section).queryByText('constancia.pdf')).toBeNull()
    expect(screen.getByText('Nada: el trámite está cerrado.')).toBeDefined()
  })

  it('says when a closed procedure has no resolution attached', async () => {
    stubApi(200, { ...detail, as_applicant: true, state: 'closed', open_step: undefined, can_act: false })
    renderPage()

    expect(await screen.findByText('El trámite se cerró sin un documento resolutivo adjunto.')).toBeDefined()
  })

  it('names the area instead of the person when the applicant cannot see who did a step', async () => {
    stubApi(200, {
      ...detail,
      as_applicant: true,
      history: [{ step: 'review', at: '2026-11-16T12:00:00Z', observation: '', declared: false, returned: false }],
    })
    renderPage()

    expect(await screen.findByText(/^Dirección de Alumnos · /)).toBeDefined()
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
          send_back: false,
          returned: false,
        },
        {
          step: 'review',
          by: { id: 's-1', name: 'Operativa Dos' },
          at: '2026-11-17T15:30:00Z',
          observation: 'Documentación completa',
          outcome: 'accept',
          declared: false,
          send_back: false,
          returned: false,
          referral: { from: 'office', to: 'committee', external_file_number: 'EX-2026-00001' },
        },
        {
          step: 'evaluate',
          by: { id: 's-2', name: '' },
          at: '2026-11-20T13:00:00Z',
          observation: 'Dictamen recibido',
          declared: true,
          send_back: false,
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
          { id: 'accept', name: 'Aceptar', next: 'evaluate', send_back: false },
          { id: 'send_back', name: 'Devolver al titular', next: 'documents', send_back: true },
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

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('offers to confirm only an open step of the staff that the person can do', async () => {
    stubPost(201, {}, referable)
    renderPage()
    expect(await screen.findByRole('heading', { level: 2, name: 'Confirmar Revisar la solicitud' })).toBeDefined()
  })

  it.each([
    ['who cannot act', { ...referable, can_act: false }],
    ['a step of the applicant', { ...referable, open_step: 'documents', state: 'with_applicant' as const }],
  ])('does not offer to confirm to %s', async (_, current) => {
    stubPost(201, {}, current)
    renderPage()
    await screen.findByRole('heading', { level: 1 })
    expect(screen.queryByRole('button', { name: 'Confirmar el paso' })).toBeNull()
  })

  it('asks for an outcome and an observation before sending', async () => {
    const fetch = stubPost(201, {}, referable)
    renderPage()

    await userEvent.click(await screen.findByRole('button', { name: 'Confirmar el paso' }))

    expect(screen.getByText('Elegí una salida.')).toBeDefined()
    expect(screen.getByText('Escribí una observación.')).toBeDefined()
    expect(await sentPosts(fetch)).toEqual([])
  })

  it('asks for the external file number only when the outcome refers the procedure', async () => {
    stubPost(201, {}, referable)
    renderPage()

    await userEvent.click(await screen.findByRole('combobox', { name: /Salida/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Devolver al titular' }))
    expect(screen.queryByLabelText('Número de expediente externo')).toBeNull()

    await userEvent.click(screen.getByRole('combobox', { name: /Salida/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Aceptar' }))
    expect(screen.getByLabelText('Número de expediente externo')).toBeDefined()
  })

  it('sends the confirmation and says it is in the history', async () => {
    const fetch = stubPost(201, {}, referable)
    renderPage()

    await userEvent.click(await screen.findByRole('combobox', { name: /Salida/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Aceptar' }))
    await userEvent.type(screen.getByRole('textbox', { name: /Observación/ }), ' Documentación completa ')
    await userEvent.type(screen.getByLabelText('Número de expediente externo'), 'EX-2026-00001')
    await userEvent.click(screen.getByRole('button', { name: 'Confirmar el paso' }))

    expect(await screen.findByText('Paso confirmado. Quedó en el historial.')).toBeDefined()
    expect(await sentPosts(fetch)).toEqual([
      {
        url: expect.stringMatching(/\/api\/procedures\/P-1\/steps\/review\/confirmation$/),
        body: { observation: 'Documentación completa', outcome: 'accept', external_file_number: 'EX-2026-00001' },
      },
    ])
    const reads = fetch.mock.calls.filter(([request]) => request.method === 'GET')
    expect(reads.length).toBeGreaterThan(1)
  })

  it('asks for the reason of a send back and says who sees it', async () => {
    const fetch = stubPost(201, {}, referable)
    renderPage()

    await userEvent.click(await screen.findByRole('combobox', { name: /Salida/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Devolver al titular' }))
    expect(screen.getByText('Devuelve el trámite a Adjuntar documentos.')).toBeDefined()
    expect(screen.getByText('Lo ve el titular.')).toBeDefined()

    await userEvent.click(screen.getByRole('button', { name: 'Confirmar el paso' }))
    expect(screen.getByText('Escribí el motivo de la devolución.')).toBeDefined()
    expect(await sentPosts(fetch)).toEqual([])
  })

  it('sends a send back with its reason', async () => {
    const fetch = stubPost(201, {}, referable)
    renderPage()

    await userEvent.click(await screen.findByRole('combobox', { name: /Salida/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Devolver al titular' }))
    await userEvent.type(screen.getByRole('textbox', { name: /Motivo de la devolución/ }), 'Falta la firma')
    await userEvent.click(screen.getByRole('button', { name: 'Confirmar el paso' }))

    expect(await screen.findByText('Paso confirmado. Quedó en el historial.')).toBeDefined()
    expect(await sentPosts(fetch)).toEqual([
      {
        url: expect.stringMatching(/\/api\/procedures\/P-1\/steps\/review\/confirmation$/),
        body: { observation: 'Falta la firma', outcome: 'send_back' },
      },
    ])
  })

  it('shows why the server did not take the confirmation', async () => {
    stubPost(409, { message: 'El paso ya no está abierto: el trámite cambió mientras tanto.' }, referable)
    renderPage()

    await userEvent.click(await screen.findByRole('combobox', { name: /Salida/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Devolver al titular' }))
    await userEvent.type(screen.getByRole('textbox', { name: /Motivo de la devolución/ }), 'Falta la firma')
    await userEvent.click(screen.getByRole('button', { name: 'Confirmar el paso' }))

    expect((await screen.findByRole('alert')).textContent).toBe(
      'El paso ya no está abierto: el trámite cambió mientras tanto.',
    )
    expect(screen.queryByText('Paso confirmado. Quedó en el historial.')).toBeNull()
  })
})

describe('ProcedurePage attachment', () => {
  const record = {
    id: 'record',
    name: 'Acta',
    required: true,
    accepted_types: ['application/pdf'],
    max_size: 1024,
    resolution: false,
  }
  const asksForRecord: Detail = {
    ...detail,
    steps: [detail.steps[0]!, { ...detail.steps[1]!, documents: [record] }],
  }

  function pdf(name: string) {
    return new File(['%PDF-1.7\n'], name, { type: 'application/pdf' })
  }

  function fileInput() {
    return document.querySelector<HTMLInputElement>('input[type="file"]')!
  }

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  const attached = {
    ...detail.steps[0]!.files[0]!,
    id: 'D-2',
    document: 'record',
    step: 'review',
    file_name: 'acta.pdf',
  }
  const withRecord = (file: typeof attached): Detail => ({
    ...asksForRecord,
    steps: [detail.steps[0]!, { ...asksForRecord.steps[1]!, files: [file] }],
  })

  function documentsSection() {
    return screen.getByRole('heading', { level: 2, name: 'Documentos del paso' }).parentElement!
  }

  it('offers the documents of the open step that have no file yet', async () => {
    stubPost(201, {}, asksForRecord)
    renderPage()

    expect(await screen.findByRole('heading', { level: 2, name: 'Documentos del paso' })).toBeDefined()
    expect(screen.getByRole('button', { name: 'Adjuntar Acta' })).toBeDefined()
  })

  it('offers to replace, not to attach, a document attached in the step', async () => {
    stubApi(200, withRecord(attached))
    renderPage()

    await screen.findByRole('heading', { level: 2, name: 'Documentos del paso' })
    expect(within(documentsSection()).getByRole('link', { name: 'Descargar acta.pdf' })).toBeDefined()
    expect(screen.getByRole('button', { name: 'Reemplazar Acta' })).toBeDefined()
    expect(screen.queryByRole('button', { name: 'Adjuntar Acta' })).toBeNull()
  })

  it('does not show the documents of the step to who cannot act', async () => {
    stubApi(200, { ...asksForRecord, can_act: false })
    renderPage()
    await screen.findAllByRole('heading', { level: 2, name: 'Historial' })
    expect(screen.queryByRole('heading', { name: 'Documentos del paso' })).toBeNull()
  })

  it('replaces a document and shows the new file as the current one', async () => {
    let current = withRecord(attached)
    const fetch = vi.fn(async (request: Request) => {
      if (request.method === 'POST') {
        current = withRecord({ ...attached, id: 'D-3', file_name: 'acta-corregida.pdf' })
        return Response.json({}, { status: 201 })
      }
      return Response.json(current)
    })
    vi.stubGlobal('fetch', fetch)
    renderPage()

    await userEvent.click(await screen.findByRole('button', { name: 'Reemplazar Acta' }))
    await userEvent.upload(fileInput(), pdf('acta-corregida.pdf'))
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar Acta' }))

    expect(await screen.findByText('Documento reemplazado. Quedó en el historial.')).toBeDefined()
    const link = await within(documentsSection()).findByRole('link', { name: 'Descargar acta-corregida.pdf' })
    expect(link.getAttribute('href')).toBe('/api/procedures/P-1/documents/D-3/content')
    expect(within(documentsSection()).queryByRole('link', { name: 'Descargar acta.pdf' })).toBeNull()
    expect(screen.getByRole('button', { name: 'Reemplazar Acta' })).toBeDefined()
  })

  it('keeps the file when replacing it is cancelled', async () => {
    stubApi(200, withRecord(attached))
    renderPage()

    await userEvent.click(await screen.findByRole('button', { name: 'Reemplazar Acta' }))
    await userEvent.click(screen.getByRole('button', { name: 'No reemplazar Acta' }))

    expect(within(documentsSection()).getByRole('link', { name: 'Descargar acta.pdf' })).toBeDefined()
    expect(screen.queryByRole('button', { name: 'Adjuntar Acta' })).toBeNull()
  })

  it('offers a document whose file was attached in another step', async () => {
    stubApi(200, withRecord({ ...attached, step: 'documents' }))
    renderPage()

    expect(await screen.findByRole('button', { name: 'Adjuntar Acta' })).toBeDefined()
  })

  it('asks for a file before sending', async () => {
    const fetch = stubPost(201, {}, asksForRecord)
    renderPage()

    await userEvent.click(await screen.findByRole('button', { name: 'Adjuntar Acta' }))

    expect(await screen.findByText('Adjuntá este documento.')).toBeDefined()
    expect(fetch.mock.calls.some(([request]) => request.method === 'POST')).toBe(false)
  })

  it('sends the file for its document and says it is in the history', async () => {
    const fetch = stubPost(201, {}, asksForRecord)
    renderPage()

    await screen.findByRole('button', { name: 'Adjuntar Acta' })
    await userEvent.upload(fileInput(), pdf('acta.pdf'))
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar Acta' }))

    expect(await screen.findByText('Documento adjuntado. Quedó en el historial.')).toBeDefined()
    const [post] = fetch.mock.calls.map(([request]) => request).filter((request) => request.method === 'POST')
    expect(post!.url).toMatch(/\/api\/procedures\/P-1\/documents$/)
    // Read as text: undici cannot parse a form with the File of jsdom, which
    // reaches it as an empty Blob without its name.
    const sent = await post!.text()
    expect(sent).toMatch(/name="document"\r\n\r\nrecord\r\n[\s\S]*name="file"/)
  })

  it('shows why the server did not take the file', async () => {
    stubPost(400, { message: 'El archivo no es de un tipo que acepte ese documento.' }, asksForRecord)
    renderPage()

    await screen.findByRole('button', { name: 'Adjuntar Acta' })
    await userEvent.upload(fileInput(), pdf('acta.pdf'))
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar Acta' }))

    expect((await screen.findByRole('alert')).textContent).toBe('El archivo no es de un tipo que acepte ese documento.')
    expect(screen.queryByText('Documento adjuntado. Quedó en el historial.')).toBeNull()
  })

  it('shows an attached document in the history, with its file', async () => {
    stubApi(200, {
      ...asksForRecord,
      history: [
        {
          step: 'review',
          by: { id: 's-1', name: 'Personal Uno' },
          at: '2026-11-17T15:00:00Z',
          observation: '',
          declared: false,
          send_back: false,
          returned: false,
          document: { id: 'D-2', document: 'record', file_name: 'acta.pdf' },
        },
      ],
    })
    renderPage()

    const title = await screen.findByText('Documento adjuntado en Revisar la solicitud')
    const entry = title.closest('.mantine-Timeline-item') as HTMLElement
    expect(within(entry).getByText('Acta')).toBeDefined()
    expect(within(entry).getByRole('button', { name: 'Ver acta.pdf' })).toBeDefined()
    const link = within(entry).getByRole('link', { name: 'Descargar acta.pdf' })
    expect(link.getAttribute('href')).toBe('/api/procedures/P-1/documents/D-2/content')
  })
})

describe('ProcedurePage send back', () => {
  const sendBack = {
    step: 'review',
    by: { id: 's-1', name: 'Operativa Dos' },
    at: '2026-11-17T15:30:00Z',
    observation: 'Falta la firma',
    outcome: 'send_back',
    declared: false,
    send_back: true,
    next: 'documents',
    returned: false,
  }
  const returned: Detail = {
    ...detail,
    state: 'with_applicant',
    open_step: 'documents',
    can_act: false,
    history: [sendBack],
  }

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('shows a send back in the history with where it went and the reason', async () => {
    stubApi(200, returned)
    renderPage()

    expect(await screen.findByText('Devuelto a Adjuntar documentos')).toBeDefined()
    expect(screen.getByText('Motivo: Falta la firma')).toBeDefined()
    expect(screen.queryByText(/^Salida:/)).toBeNull()
  })

  it('tells the applicant why the procedure came back', async () => {
    stubApi(200, returned)
    renderPage()

    const alert = (await screen.findByText('Trámite devuelto')).closest('[role="alert"]') as HTMLElement
    expect(within(alert).getByText('Volvió al paso Adjuntar documentos. Motivo:')).toBeDefined()
    expect(within(alert).getByText('Falta la firma')).toBeDefined()
  })

  it('does not say it came back once the procedure moved on', async () => {
    stubApi(200, { ...returned, open_step: 'review', state: 'in_office', history: [sendBack] })
    renderPage()

    await screen.findByRole('heading', { level: 1 })
    expect(screen.queryByText('Trámite devuelto')).toBeNull()
  })

  it('does not say it came back when the send back stays in an area', async () => {
    stubApi(200, { ...returned, open_step: 'review', history: [{ ...sendBack, next: 'review' }] })
    renderPage()

    await screen.findByRole('heading', { level: 1 })
    expect(screen.queryByText('Trámite devuelto')).toBeNull()
  })
})

describe('ProcedurePage correction request', () => {
  const asksForSubjects: Detail = {
    ...detail,
    steps: [
      {
        ...detail.steps[0]!,
        fields: [{ id: 'subjects', name: 'Materias', required: true, type: 'text', options: [] }],
      },
      detail.steps[1]!,
    ],
  }
  const request = {
    step: 'review',
    by: { id: 's-1', name: 'Operativo Uno' },
    at: '2026-11-17T15:00:00Z',
    reason: 'Falta el código de cada materia.',
    fields: ['subjects'],
    documents: ['statement'],
  }
  const awaiting: Detail = { ...asksForSubjects, state: 'with_applicant', correction_requests: [request] }

  async function openForm() {
    await userEvent.click(await screen.findByRole('button', { name: 'Pedir una subsanación' }))
  }

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it.each([
    ['who cannot act', { ...asksForSubjects, can_act: false }],
    ['a step of the applicant', { ...asksForSubjects, open_step: 'documents', state: 'with_applicant' as const }],
  ])('does not offer to request a correction to %s', async (_, current) => {
    stubPost(201, {}, current)
    renderPage()
    await screen.findByRole('heading', { level: 1 })
    expect(screen.queryByRole('button', { name: 'Pedir una subsanación' })).toBeNull()
  })

  it('asks what is missing and why before sending', async () => {
    const fetch = stubPost(201, {}, asksForSubjects)
    renderPage()

    await openForm()
    await userEvent.click(screen.getByRole('button', { name: 'Pedir la subsanación' }))

    expect(screen.getByText('Elegí qué falta.')).toBeDefined()
    expect(screen.getByText('Escribí el motivo.')).toBeDefined()
    expect(await sentPosts(fetch)).toEqual([])
  })

  it('sends the fields and documents asked for with the reason and says it is in the history', async () => {
    const fetch = stubPost(201, {}, asksForSubjects)
    renderPage()

    await openForm()
    await userEvent.click(screen.getByRole('combobox', { name: /Qué falta/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Materias' }))
    await userEvent.click(screen.getByRole('option', { name: 'Constancia' }))
    await userEvent.type(screen.getByRole('textbox', { name: /Motivo/ }), ' Falta el código de cada materia. ')
    await userEvent.click(screen.getByRole('button', { name: 'Pedir la subsanación' }))

    expect(await screen.findByText('Subsanación pedida. Quedó en el historial.')).toBeDefined()
    expect(await sentPosts(fetch)).toEqual([
      {
        url: expect.stringMatching(/\/api\/procedures\/P-1\/steps\/review\/correction-requests$/),
        body: { reason: 'Falta el código de cada materia.', fields: ['subjects'], documents: ['statement'] },
      },
    ])
  })

  it('offers only what the steps of the applicant ask for', async () => {
    const ruling = {
      id: 'ruling',
      name: 'Dictamen',
      required: true,
      accepted_types: ['application/pdf'],
      max_size: 1024,
      resolution: false,
    }
    stubPost(
      201,
      {},
      {
        ...asksForSubjects,
        steps: [asksForSubjects.steps[0]!, { ...asksForSubjects.steps[1]!, documents: [ruling] }],
      },
    )
    renderPage()

    await openForm()
    await userEvent.click(screen.getByRole('combobox', { name: /Qué falta/ }))

    expect(screen.getByRole('option', { name: 'Constancia' })).toBeDefined()
    expect(screen.queryByRole('option', { name: 'Dictamen' })).toBeNull()
  })

  it('shows why the server did not take the request', async () => {
    stubPost(409, { message: 'El paso ya no está abierto: el trámite cambió mientras tanto.' }, asksForSubjects)
    renderPage()

    await openForm()
    await userEvent.click(screen.getByRole('combobox', { name: /Qué falta/ }))
    await userEvent.click(screen.getByRole('option', { name: 'Materias' }))
    await userEvent.type(screen.getByRole('textbox', { name: /Motivo/ }), 'Falta el código')
    await userEvent.click(screen.getByRole('button', { name: 'Pedir la subsanación' }))

    expect((await screen.findByRole('alert')).textContent).toBe(
      'El paso ya no está abierto: el trámite cambió mientras tanto.',
    )
    expect(screen.queryByText('Subsanación pedida. Quedó en el historial.')).toBeNull()
  })

  it('shows the applicant what is pending with the area that asked and what is expected of them', async () => {
    const { by: _, ...withoutWho } = request
    stubPost(201, {}, { ...awaiting, as_applicant: true, can_act: false, correction_requests: [withoutWho] })
    renderPage()

    expect(await screen.findByText('Subsanación pendiente')).toBeDefined()
    expect(screen.getByText(/^Dirección de Alumnos · /)).toBeDefined()
    expect(screen.queryByText(/Operativo Uno/)).toBeNull()
    expect(screen.getByText('Lo que pide la subsanación, a cargo del titular.')).toBeDefined()
  })

  it('shows what is pending and does not offer to confirm the step until the applicant answers', async () => {
    stubPost(201, {}, awaiting)
    renderPage()

    expect(await screen.findByText('Subsanación pendiente')).toBeDefined()
    expect(screen.getByText('Se pide: Materias, Constancia')).toBeDefined()
    expect(screen.getByText('Falta el código de cada materia.')).toBeDefined()
    expect(screen.queryByRole('button', { name: 'Confirmar el paso' })).toBeNull()
    expect(screen.queryByText('Acción requerida')).toBeNull()
    expect(screen.getByRole('button', { name: 'Pedir una subsanación' })).toBeDefined()
  })

  it('shows a correction request in the history with what it asks for and why', async () => {
    stubPost(
      201,
      {},
      {
        ...awaiting,
        history: [
          {
            step: 'review',
            by: request.by,
            at: request.at,
            observation: request.reason,
            declared: false,
            send_back: false,
            returned: false,
            correction_request: { fields: ['subjects'], documents: [] },
          },
        ],
      },
    )
    renderPage()

    expect(await screen.findByText('Subsanación pedida en Revisar la solicitud')).toBeDefined()
    expect(screen.getByText('Se pide: Materias')).toBeDefined()
    expect(screen.getByText('Motivo: Falta el código de cada materia.')).toBeDefined()
  })
})
