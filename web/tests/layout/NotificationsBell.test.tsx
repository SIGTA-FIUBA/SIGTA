import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { NotificationsBell } from '../../src/layout/NotificationsBell'
import { render } from '../test-utils/render'

const mine = (id: string, name: string, moved: string, extra: object = {}) => ({
  id,
  procedure_type: 1,
  procedure_type_name: name,
  step_name: 'Revisión de la solicitud',
  area: { id: 'office', name: 'Secretaría' },
  state: 'in_office',
  correction_pending: false,
  started_at: '2026-11-01T12:00:00Z',
  last_moved_at: moved,
  resolutions: [],
  ...extra,
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
// started and /procedures/P-2 the detail of the one with a correction.
function stubApi(procedures: unknown[], status = 200) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async (request: Request) => {
      if (status !== 200) {
        return Response.json({ message: 'error' }, { status })
      }
      const path = new URL(request.url).pathname
      if (path === '/api/me/procedures') {
        return Response.json(procedures)
      }
      if (path === '/api/procedures/P-2') {
        return Response.json(awaiting)
      }
      return Response.json({ message: 'No existe' }, { status: 404 })
    }),
  )
}

function renderBell() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <NotificationsBell />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('NotificationsBell', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('opens what is required and the latest moves of the procedures in progress', async () => {
    stubApi([
      mine('P-1', 'Equivalencia entre asignaturas', '2026-11-10T10:00:00Z'),
      mine('P-2', 'Reconocimiento de créditos', '2026-11-17T15:00:00Z', {
        state: 'with_applicant',
        correction_pending: true,
      }),
      mine('P-3', 'Pase de carrera', '2026-11-18T10:00:00Z', { state: 'closed' }),
    ])
    renderBell()

    const bell = await screen.findByRole('button', { name: 'Notificaciones, con acción requerida' })
    await userEvent.click(bell)

    const dropdown = await screen.findByRole('dialog')
    expect(within(dropdown).getByRole('link', { name: 'Materias aprobadas' })).toHaveAttribute(
      'href',
      '/procedures/P-2',
    )
    const moves = within(dropdown)
      .getAllByRole('link')
      .slice(1)
      .map((link) => link.textContent)
    // The latest first, and the closed ones are in the history.
    expect(moves).toEqual(['Reconocimiento de créditos', 'Equivalencia entre asignaturas'])
    expect(within(dropdown).getByText('Revisión de la solicitud, En Secretaría')).toBeInTheDocument()
  })

  it('says so when nothing changed', async () => {
    stubApi([])
    renderBell()

    await userEvent.click(await screen.findByRole('button', { name: 'Notificaciones' }))

    expect(await screen.findByText('Sin novedades')).toBeInTheDocument()
  })

  it('closes when a notification is followed', async () => {
    stubApi([mine('P-1', 'Equivalencia entre asignaturas', '2026-11-10T10:00:00Z')])
    renderBell()

    await userEvent.click(await screen.findByRole('button', { name: 'Notificaciones' }))
    await userEvent.click(await screen.findByRole('link', { name: 'Equivalencia entre asignaturas' }))

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('marks a procedure waiting on the person for a step of theirs, by that step', async () => {
    stubApi([
      mine('P-4', 'Pase de carrera', '2026-11-18T10:00:00Z', {
        state: 'with_applicant',
        step_name: 'Aceptar la resolución',
      }),
    ])
    renderBell()

    await userEvent.click(await screen.findByRole('button', { name: 'Notificaciones, con acción requerida' }))

    const dropdown = await screen.findByRole('dialog')
    expect(within(dropdown).getAllByRole('link', { name: 'Aceptar la resolución' })[0]).toHaveAttribute(
      'href',
      '/procedures/P-4',
    )
  })

  it('says when the news could not be loaded instead of saying there are none', async () => {
    stubApi([], 500)
    renderBell()

    await userEvent.click(screen.getByRole('button', { name: 'Notificaciones' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudieron cargar las novedades.')
    expect(screen.queryByText('Sin novedades')).not.toBeInTheDocument()
  })

  it('closes with Escape and gives the focus back to the bell', async () => {
    stubApi([])
    renderBell()
    const bell = await screen.findByRole('button', { name: 'Notificaciones' })

    await userEvent.click(bell)
    expect(bell).toHaveAttribute('aria-expanded', 'true')
    await userEvent.keyboard('{Escape}')

    expect(bell).toHaveAttribute('aria-expanded', 'false')
    expect(bell).toHaveFocus()
  })
})
