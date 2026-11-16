import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
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
