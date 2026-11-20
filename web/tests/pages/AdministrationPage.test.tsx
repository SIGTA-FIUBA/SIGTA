import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { AdministrationPage } from '../../src/pages/AdministrationPage'
import { render } from '../test-utils/render'

const secretariat = {
  id: 'secretariat',
  name: 'Secretaría',
  system: true,
  permissions: [{ id: 'publish_procedure_types', name: 'Publicar tipos de trámite', system: true }],
  roles: [],
}
const office = {
  id: 'office',
  name: 'Dirección de Alumnos',
  system: false,
  permissions: [
    { id: 'review', name: 'Revisar', system: false },
    { id: 'approve', name: 'Aprobar', system: false },
  ],
  roles: [{ id: 'clerk', name: 'Administrativo', permissions: ['review'] }],
}
const person = { id: 'u2', name: 'Persona Dos', email: 'persona.dos@fi.uba.ar' }
const membership = {
  id: 'm1',
  person,
  roles: ['clerk'],
  permissions: [],
  scope: [],
  from: '2026-11-01T12:00:00Z',
  granted_by: { id: 'u1', name: 'Persona Uno' },
}
const administrator = {
  id: 'a1',
  person: { id: 'u1', name: 'Persona Uno', email: 'persona.uno@fi.uba.ar' },
  from: '2026-10-01T12:00:00Z',
}

type Answer = { status: number; body?: unknown }

// stubApi plays the API: GETs answer from reads, and every other request
// from changes, keyed by method and path, or 204 when it is not there.
function stubApi(reads: Record<string, Answer>, changes: Record<string, Answer> = {}) {
  const fetch = vi.fn(async (request: Request) => {
    const url = new URL(request.url)
    const key = `${request.method} ${url.pathname.replace(/^\/api/, '')}`
    const answer = request.method === 'GET' ? reads[url.pathname.replace(/^\/api/, '')] : changes[key]
    if (answer === undefined) {
      return request.method === 'GET'
        ? Response.json({ message: 'No existe.' }, { status: 404 })
        : new Response(null, { status: 204 })
    }
    return answer.body === undefined
      ? new Response(null, { status: answer.status })
      : Response.json(answer.body, { status: answer.status })
  })
  vi.stubGlobal('fetch', fetch)
  return fetch
}

const reads = {
  '/areas': { status: 200, body: [secretariat, office] },
  '/areas/secretariat/memberships': { status: 200, body: [] },
  '/areas/office/memberships': { status: 200, body: [membership] },
  '/administrators': { status: 200, body: [administrator] },
  '/procedure-types': { status: 200, body: [] },
  '/people': { status: 200, body: person },
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
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <AdministrationPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const area = (name: string) => screen.findByRole('region', { name })

describe('AdministrationPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('lists the areas with their permissions, roles and memberships', async () => {
    stubApi(reads)
    renderPage()

    const secretariatPanel = await area('Secretaría')
    expect(within(secretariatPanel).getByText('Viene con SIGTA')).toBeVisible()
    expect(within(secretariatPanel).getByText('De plataforma')).toBeVisible()
    expect(within(secretariatPanel).queryByRole('button', { name: /Dar de baja/ })).not.toBeInTheDocument()
    const officePanel = await area('Dirección de Alumnos')
    expect(within(officePanel).getByText('review')).toBeVisible()
    expect(within(officePanel).getByText('Administrativo')).toBeVisible()
    expect(await within(officePanel).findByText('Persona Dos')).toBeVisible()
    expect(within(officePanel).getByText('Roles: Administrativo')).toBeVisible()
    expect(within(officePanel).getByText('Alcance: todos los tipos de trámite')).toBeVisible()
    expect(
      within(officePanel).getByText(/Desde el 01\/11\/2026, sin fecha de fin\. Otorgada por Persona Uno\./),
    ).toBeVisible()
  })

  it('creates an area', async () => {
    const fetch = stubApi(reads, { 'POST /areas': { status: 201, body: { ...office, id: 'board' } } })
    renderPage()

    await area('Secretaría')
    await userEvent.click(screen.getByRole('button', { name: 'Crear un área' }))
    const form = screen.getByRole('heading', { name: 'Crear un área' }).closest('form')!
    await userEvent.type(within(form).getByLabelText(/Identificador/), 'board')
    await userEvent.type(within(form).getByLabelText('Nombre'), 'Consejo Directivo')
    await userEvent.click(within(form).getByRole('button', { name: 'Crear área' }))

    await vi.waitFor(async () =>
      expect(await sent(fetch)).toEqual([{ request: 'POST /areas', body: { id: 'board', name: 'Consejo Directivo' } }]),
    )
  })

  it('opens a form in place of its button and gives the focus back on cancel', async () => {
    stubApi(reads)
    renderPage()

    const panel = await area('Dirección de Alumnos')
    await userEvent.click(within(panel).getByRole('button', { name: 'Agregar un permiso' }))
    expect(within(panel).getByLabelText('Nombre')).toHaveFocus()
    await userEvent.click(within(panel).getByRole('button', { name: 'Cancelar' }))
    expect(within(panel).getByRole('button', { name: 'Agregar un permiso' })).toHaveFocus()
  })

  it('says why the API refused to delete an area', async () => {
    const message = 'El área tiene membresías vigentes o futuras: terminalas antes de darla de baja.'
    stubApi(reads, { 'DELETE /areas/office': { status: 409, body: { message } } })
    renderPage()

    const panel = await area('Dirección de Alumnos')
    await userEvent.click(within(panel).getByRole('button', { name: 'Dar de baja el área Dirección de Alumnos' }))
    await userEvent.click(within(panel).getByRole('button', { name: 'Dar de baja' }))

    expect(await within(panel).findByRole('alert')).toHaveTextContent(message)
  })

  it('grants a membership to a person found by email', async () => {
    const fetch = stubApi(reads, { 'POST /memberships': { status: 201, body: {} } })
    renderPage()

    const panel = await area('Dirección de Alumnos')
    await userEvent.click(within(panel).getByRole('button', { name: 'Otorgar una membresía' }))
    await userEvent.type(within(panel).getByLabelText(/Correo de la persona/), 'persona.dos@fi.uba.ar')
    await userEvent.click(within(panel).getByRole('button', { name: 'Buscar' }))
    await within(panel).findByRole('button', { name: 'Elegir otra persona' })
    await userEvent.click(within(panel).getByRole('combobox', { name: /Permisos sueltos/ }))
    await userEvent.click(await screen.findByRole('option', { name: 'Aprobar' }))
    await userEvent.type(within(panel).getByLabelText(/Hasta/), '2027-03-31')
    await userEvent.click(within(panel).getByRole('button', { name: 'Otorgar membresía' }))

    await vi.waitFor(async () =>
      expect(await sent(fetch)).toEqual([
        {
          request: 'POST /memberships',
          body: {
            user_id: 'u2',
            area: 'office',
            roles: [],
            permissions: ['approve'],
            scope: [],
            until: '2027-04-01T02:59:59.000Z',
          },
        },
      ]),
    )
  })

  it('says when nobody registered with the email', async () => {
    stubApi({ ...reads, '/people': { status: 404, body: { message: 'No existe.' } } })
    renderPage()

    const panel = await area('Dirección de Alumnos')
    await userEvent.click(within(panel).getByRole('button', { name: 'Otorgar una membresía' }))
    await userEvent.type(within(panel).getByLabelText(/Correo de la persona/), 'nadie@fi.uba.ar')
    await userEvent.click(within(panel).getByRole('button', { name: 'Buscar' }))

    expect(await within(panel).findByRole('alert')).toHaveTextContent('No hay una persona registrada con ese correo')
  })

  it('changes the permissions of a membership', async () => {
    const fetch = stubApi(reads, { 'PUT /memberships/m1/permissions': { status: 200, body: {} } })
    renderPage()

    const panel = await area('Dirección de Alumnos')
    await userEvent.click(await within(panel).findByRole('button', { name: 'Cambiar los permisos de Persona Dos' }))
    await userEvent.click(within(panel).getByRole('combobox', { name: /Permisos sueltos/ }))
    await userEvent.click(await screen.findByRole('option', { name: 'Aprobar' }))
    await userEvent.click(within(panel).getByRole('button', { name: 'Guardar permisos' }))

    await vi.waitFor(async () =>
      expect(await sent(fetch)).toEqual([
        { request: 'PUT /memberships/m1/permissions', body: { roles: ['clerk'], permissions: ['approve'] } },
      ]),
    )
  })

  it('ends a membership after asking again', async () => {
    const fetch = stubApi(reads)
    renderPage()

    const panel = await area('Dirección de Alumnos')
    await userEvent.click(await within(panel).findByRole('button', { name: 'Terminar la membresía de Persona Dos' }))
    expect(await sent(fetch)).toEqual([])
    await userEvent.click(within(panel).getByRole('button', { name: 'Terminar membresía' }))

    await vi.waitFor(async () =>
      expect(await sent(fetch)).toEqual([{ request: 'DELETE /memberships/m1', body: undefined }]),
    )
  })

  it('lists the administrators and grants the administration', async () => {
    const fetch = stubApi(reads, { 'POST /administrators': { status: 201, body: {} } })
    renderPage()

    await userEvent.click(await screen.findByRole('tab', { name: 'Personas administradoras' }))
    expect(await screen.findByText('Persona Uno')).toBeVisible()
    expect(screen.getByText(/Otorgada por consola\./)).toBeVisible()
    // The last administrator cannot be removed: SIGTA never runs out of them.
    expect(screen.getByText('Única persona administradora')).toBeVisible()
    expect(screen.queryByRole('button', { name: /Quitarle la administración/ })).not.toBeInTheDocument()

    await userEvent.type(screen.getByLabelText(/Correo de la persona/), 'persona.dos@fi.uba.ar')
    await userEvent.click(screen.getByRole('button', { name: 'Buscar' }))
    await userEvent.click(await screen.findByRole('button', { name: 'Hacer administradora' }))

    await vi.waitFor(async () =>
      expect(await sent(fetch)).toEqual([{ request: 'POST /administrators', body: { user_id: 'u2' } }]),
    )
  })

  it('removes an administrator when there is another', async () => {
    const other = { ...administrator, id: 'a2', person, granted_by: { id: 'u1', name: 'Persona Uno' } }
    const fetch = stubApi({ ...reads, '/administrators': { status: 200, body: [administrator, other] } })
    renderPage()

    await userEvent.click(await screen.findByRole('tab', { name: 'Personas administradoras' }))
    await userEvent.click(await screen.findByRole('button', { name: 'Quitarle la administración a Persona Dos' }))
    await userEvent.click(screen.getByRole('button', { name: 'Quitar administración' }))

    await vi.waitFor(async () =>
      expect(await sent(fetch)).toEqual([{ request: 'DELETE /administrators/a2', body: undefined }]),
    )
  })

  it('tells someone who is not an administrator that the page is not for them', async () => {
    stubApi({
      ...reads,
      '/areas': { status: 403, body: { message: 'Solo una persona administradora puede hacer esto.' } },
    })
    renderPage()

    expect(await screen.findByRole('alert')).toHaveTextContent('Solo una persona administradora puede ver esta página.')
  })

  it('says when the areas fail to load', async () => {
    stubApi({ ...reads, '/areas': { status: 500, body: { message: 'Ocurrió un error inesperado.' } } })
    renderPage()

    expect(await screen.findByRole('alert')).toHaveTextContent('No pudimos cargar las áreas')
  })
})
