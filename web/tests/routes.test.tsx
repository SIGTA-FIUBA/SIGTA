import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { routes } from '../src/routes'
import { render } from './render'

const user = { id: 'u1', email: 'titular@fi.uba.ar', name: 'Ana Titular', role: 'applicant' }

// stubApi plays the API: /me answers while the session is open, and a
// successful /auth/logout closes it.
function stubApi({ session, logoutStatus = 204 }: { session: boolean; logoutStatus?: number }) {
  let open = session
  const fetch = vi.fn(async (request: Request) => {
    if (request.url.endsWith('/auth/logout')) {
      if (logoutStatus === 204) {
        open = false
      }
      return new Response(null, { status: logoutStatus })
    }
    return open ? Response.json(user) : Response.json({ message: 'Sin sesión' }, { status: 401 })
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

describe('routes', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('sends a visitor without a session to the login page', async () => {
    stubApi({ session: false })
    const router = renderAt('/')

    expect(await screen.findByRole('link', { name: 'Iniciar sesión con Google' })).toBeVisible()
    expect(router.state.location.pathname).toBe('/login')
  })

  it('shows the catalog to a visitor without a session', async () => {
    stubApi({ session: false })
    const router = renderAt('/catalog')

    expect(await screen.findByRole('heading', { name: 'Catálogo de trámites' })).toBeVisible()
    expect(router.state.location.pathname).toBe('/catalog')
    // A visitor gets the frame with a way to log in, and no menu.
    expect(screen.getByRole('link', { name: 'Iniciar sesión' })).toHaveAttribute('href', '/login')
    expect(screen.getByRole('link', { name: 'Volver al inicio' })).toHaveAttribute('href', '/login')
    expect(screen.getByRole('link', { name: 'SIGTA, ir al inicio' })).toHaveAttribute('href', '/')
    expect(screen.queryByRole('navigation', { name: 'Menú principal' })).not.toBeInTheDocument()
  })

  it('shows the home page with the account of the person with a session', async () => {
    stubApi({ session: true })
    const router = renderAt('/')

    expect(
      await screen.findByRole('heading', { name: 'Sistema Integral de Gestión de Trámites Académicos' }),
    ).toBeVisible()
    expect(screen.getByRole('button', { name: 'Cuenta de Ana Titular' })).toHaveTextContent('AT')
    expect(router.state.location.pathname).toBe('/')
  })

  it('frames the pages with the faculty logo and the main menu', async () => {
    stubApi({ session: true })
    renderAt('/')

    await screen.findByRole('heading', { name: 'Sistema Integral de Gestión de Trámites Académicos' })
    expect(screen.getByRole('link', { name: 'SIGTA, ir al inicio' })).toHaveAttribute('href', '/')
    const menu = screen.getByRole('navigation', { name: 'Menú principal' })
    expect(within(menu).getByRole('link', { name: 'Inicio' })).toHaveAttribute('href', '/')
    expect(within(menu).getByRole('link', { name: 'Catálogo de trámites' })).toHaveAttribute('href', '/catalog')
    // The applicant has no inbox, and the menu says who has the session.
    expect(within(menu).queryByRole('link', { name: /Bandeja de trámites/ })).not.toBeInTheDocument()
    expect(within(menu).getByText('Titular')).toBeVisible()
  })

  it('keeps the menu on the catalog for whoever has a session, to go back', async () => {
    stubApi({ session: true })
    renderAt('/catalog')

    await screen.findByRole('heading', { name: 'Catálogo de trámites' })
    const menu = screen.getByRole('navigation', { name: 'Menú principal' })
    expect(within(menu).getByRole('link', { name: 'Inicio' })).toHaveAttribute('href', '/')
    expect(within(menu).getByRole('link', { name: 'Catálogo de trámites' })).toHaveAttribute('data-active', 'true')
  })

  it('sends someone with a session from the login page to the home page', async () => {
    stubApi({ session: true })
    const router = renderAt('/login')

    expect(
      await screen.findByRole('heading', { name: 'Sistema Integral de Gestión de Trámites Académicos' }),
    ).toBeVisible()
    expect(router.state.location.pathname).toBe('/')
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
