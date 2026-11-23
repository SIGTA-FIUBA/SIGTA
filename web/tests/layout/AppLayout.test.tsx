import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { routes } from '../../src/routes'
import { render } from '../test-utils/render'

const user = { id: 'u1', email: 'titular@fi.uba.ar', name: 'Ana Titular', staff: false, administrator: false }

const nothingToEdit = { can_create: false, procedure_types: [], areas: [] }

// stubApi plays the API: /me answers while the session is open, with no
// procedures of its own nor in the inbox, editable says what procedure types
// the person can edit, and a successful /auth/logout closes it.
function stubApi({
  session,
  logoutStatus = 204,
  account = user,
  editable = nothingToEdit,
}: {
  session: boolean
  logoutStatus?: number
  account?: typeof user
  editable?: unknown
}) {
  let open = session
  const fetch = vi.fn(async (request: Request) => {
    if (request.url.endsWith('/auth/logout')) {
      if (logoutStatus === 204) {
        open = false
      }
      return new Response(null, { status: logoutStatus })
    }
    if (!open) {
      return Response.json({ message: 'Sin sesión' }, { status: 401 })
    }
    if (request.url.endsWith('/editable-procedure-types')) {
      return Response.json(editable)
    }
    if (new URL(request.url).pathname === '/api/procedures') {
      return Response.json({ items: [] })
    }
    if (new URL(request.url).pathname === '/api/procedure-types') {
      return Response.json([])
    }
    return Response.json(request.url.endsWith('/me/procedures') ? [] : account)
  })
  vi.stubGlobal('fetch', fetch)
  return fetch
}

function renderAt(path: string) {
  const router = createMemoryRouter(routes, { initialEntries: [path] })
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
  return router
}

describe('AppLayout', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('frames the pages with the SIGTA logo and the main menu', async () => {
    stubApi({ session: true })
    renderAt('/')

    await screen.findByRole('heading', { level: 1, name: 'Mis trámites' })
    const menu = screen.getByRole('navigation', { name: 'Menú principal' })
    expect(within(menu).getByRole('link', { name: 'SIGTA, ir al inicio' })).toHaveAttribute('href', '/')
    // The applicant has their procedures on the home page, named after them.
    expect(within(menu).getByRole('link', { name: 'Mis trámites' })).toHaveAttribute('href', '/')
    expect(within(menu).queryByRole('link', { name: 'Inicio' })).not.toBeInTheDocument()
    expect(within(menu).getByRole('link', { name: 'Catálogo de trámites' })).toHaveAttribute('href', '/catalog')
    expect(within(menu).getByRole('button', { name: 'Buscar' })).toBeVisible()
    expect(within(menu).getByRole('link', { name: 'Mis trámites' })).toHaveAttribute('aria-current', 'page')
    expect(within(menu).getByRole('link', { name: 'Acerca de SIGTA' })).toHaveAttribute('href', '/about')
    // The applicant has no inbox, and the account closes the menu.
    expect(within(menu).queryByRole('link', { name: /Bandeja de trámites/ })).not.toBeInTheDocument()
    expect(
      within(within(menu).getByRole('button', { name: 'Cuenta de Ana Titular' })).getByText('Titular'),
    ).toBeVisible()
    expect(within(menu).queryByRole('link', { name: 'Administración' })).not.toBeInTheDocument()
  })

  it('takes an administrator to the administration', async () => {
    stubApi({ session: true, account: { ...user, administrator: true } })
    renderAt('/')

    const menu = await screen.findByRole('navigation', { name: 'Menú principal' })
    expect(within(menu).getByRole('link', { name: 'Administración' })).toHaveAttribute('href', '/administration')
  })

  it('offers the procedure types to the staff who can edit or approve one', async () => {
    stubApi({ session: true, account: { ...user, staff: true }, editable: { ...nothingToEdit, can_create: true } })
    renderAt('/')

    const menu = await screen.findByRole('navigation', { name: 'Menú principal' })
    expect(await within(menu).findByRole('link', { name: 'Tipos de trámite' })).toHaveAttribute(
      'href',
      '/procedure-types',
    )
  })

  it('reads again what the person can edit when the list of procedure types opens', async () => {
    const fetch = stubApi({
      session: true,
      account: { ...user, staff: true },
      editable: { ...nothingToEdit, can_create: true },
    })
    const router = renderAt('/')
    const menu = await screen.findByRole('navigation', { name: 'Menú principal' })
    await within(menu).findByRole('link', { name: 'Tipos de trámite' })

    await router.navigate('/procedure-types')

    const reads = () => fetch.mock.calls.filter(([request]) => request.url.endsWith('/editable-procedure-types'))
    await vi.waitFor(() => expect(reads()).toHaveLength(2))
  })

  it('does not offer the procedure types to the staff who can edit none', async () => {
    stubApi({ session: true, account: { ...user, staff: true } })
    renderAt('/')

    const menu = await screen.findByRole('navigation', { name: 'Menú principal' })
    expect(await within(menu).findByRole('link', { name: /Bandeja de trámites/ })).toBeVisible()
    expect(within(menu).queryByRole('link', { name: 'Tipos de trámite' })).not.toBeInTheDocument()
  })

  it('keeps the menu on the catalog for whoever has a session, to go back', async () => {
    stubApi({ session: true })
    renderAt('/catalog')

    await screen.findByRole('heading', { name: 'Catálogo de trámites' })
    const menu = screen.getByRole('navigation', { name: 'Menú principal' })
    expect(within(menu).getByRole('link', { name: 'Mis trámites' })).toHaveAttribute('href', '/')
    expect(within(menu).getByRole('link', { name: 'Catálogo de trámites' })).toHaveAttribute('data-active', 'true')
  })

  it('fixes the search and the start of a procedure at the top for an applicant', async () => {
    stubApi({ session: true })
    renderAt('/catalog')

    await screen.findByRole('heading', { name: 'Catálogo de trámites' })
    const menu = screen.getByRole('navigation', { name: 'Menú principal' })
    const starts = screen.getAllByRole('link', { name: 'Iniciar trámite' })
    expect(starts.length).toBeGreaterThan(0)
    for (const start of starts) {
      expect(start).toHaveAttribute('href', '/new')
    }
    const searches = screen.getAllByRole('button', { name: /^Buscar/ }).filter((button) => !menu.contains(button))
    expect(searches.length).toBeGreaterThan(0)
  })

  it('does not offer to start a procedure on the page that starts one', async () => {
    stubApi({ session: true })
    renderAt('/new')

    await screen.findByRole('heading', { level: 1, name: 'Iniciar trámite' })
    expect(screen.queryByRole('link', { name: 'Iniciar trámite' })).not.toBeInTheDocument()
  })

  it('leaves the start of a procedure out of the pages of the staff', async () => {
    stubApi({ session: true, account: { ...user, staff: true } })
    renderAt('/')

    await screen.findByRole('heading', { level: 1, name: 'Inicio' })
    expect(screen.queryByRole('link', { name: 'Iniciar trámite' })).not.toBeInTheDocument()
  })

  it('shows the name and email in the account menu', async () => {
    stubApi({ session: true })
    renderAt('/')

    await userEvent.click(await screen.findByRole('button', { name: 'Cuenta de Ana Titular' }))

    const accountMenu = screen.getByRole('menu')
    expect(within(accountMenu).getByText('Ana Titular')).toBeVisible()
    expect(within(accountMenu).getByText('titular@fi.uba.ar')).toBeVisible()
  })

  it('closes the session and goes to the login page', async () => {
    const fetch = stubApi({ session: true })
    const router = renderAt('/')

    await userEvent.click(await screen.findByRole('button', { name: 'Cuenta de Ana Titular' }))
    await userEvent.click(screen.getByRole('menuitem', { name: 'Cerrar sesión' }))

    expect(await screen.findByRole('link', { name: 'Iniciar sesión' })).toBeVisible()
    expect(router.state.location.pathname).toBe('/login')
    expect(fetch).toHaveBeenCalledWith(
      expect.objectContaining({ method: 'POST', url: expect.stringMatching(/\/auth\/logout$/) }),
    )
  })

  it('stays on the page and says so when the session cannot be closed', async () => {
    stubApi({ session: true, logoutStatus: 500 })
    const router = renderAt('/')

    await userEvent.click(await screen.findByRole('button', { name: 'Cuenta de Ana Titular' }))
    await userEvent.click(screen.getByRole('menuitem', { name: 'Cerrar sesión' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudo cerrar la sesión.')
    expect(router.state.location.pathname).toBe('/')
  })
})
