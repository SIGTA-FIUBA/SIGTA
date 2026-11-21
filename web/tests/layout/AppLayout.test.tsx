import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { routes } from '../../src/routes'
import { render } from '../test-utils/render'

const user = { id: 'u1', email: 'titular@fi.uba.ar', name: 'Ana Titular', staff: false, administrator: false }

// stubApi plays the API: /me answers while the session is open, with no
// procedures of its own, and a successful /auth/logout closes it.
function stubApi({
  session,
  logoutStatus = 204,
  account = user,
}: {
  session: boolean
  logoutStatus?: number
  account?: typeof user
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

    await screen.findByRole('heading', { level: 1, name: 'Hola, Ana' })
    const menu = screen.getByRole('navigation', { name: 'Menú principal' })
    expect(within(menu).getByRole('link', { name: 'SIGTA, ir al inicio' })).toHaveAttribute('href', '/')
    expect(within(menu).getByRole('link', { name: 'Inicio' })).toHaveAttribute('href', '/')
    expect(within(menu).getByRole('link', { name: 'Catálogo de trámites' })).toHaveAttribute('href', '/catalog')
    expect(within(menu).getByRole('link', { name: 'Inicio' })).toHaveAttribute('aria-current', 'page')
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

  it('keeps the menu on the catalog for whoever has a session, to go back', async () => {
    stubApi({ session: true })
    renderAt('/catalog')

    await screen.findByRole('heading', { name: 'Catálogo de trámites' })
    const menu = screen.getByRole('navigation', { name: 'Menú principal' })
    expect(within(menu).getByRole('link', { name: 'Inicio' })).toHaveAttribute('href', '/')
    expect(within(menu).getByRole('link', { name: 'Catálogo de trámites' })).toHaveAttribute('data-active', 'true')
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

    expect(await screen.findByRole('link', { name: 'Iniciar sesión con Google' })).toBeVisible()
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
