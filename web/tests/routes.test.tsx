import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { routes } from '../src/routes'
import { render } from './test-utils/render'

const user = { id: 'u1', email: 'titular@fi.uba.ar', name: 'Ana Titular', staff: false, administrator: false }

// stubApi plays the API: /me answers only while the session is open.
function stubApi({ session }: { session: boolean }) {
  const fetch = vi.fn(async (_request: Request) =>
    session ? Response.json(user) : Response.json({ message: 'Sin sesión' }, { status: 401 }),
  )
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

    expect(await screen.findByRole('link', { name: 'Iniciar sesión' })).toBeVisible()
    expect(router.state.location.pathname).toBe('/login')
  })

  it('shows the catalog to a visitor without a session', async () => {
    stubApi({ session: false })
    const router = renderAt('/catalog')

    expect(await screen.findByRole('heading', { name: 'Catálogo de trámites' })).toBeVisible()
    expect(router.state.location.pathname).toBe('/catalog')
    // A visitor gets the header of the public pages, with a way to log in,
    // and no menu.
    expect(screen.getByRole('link', { name: 'Iniciar sesión' })).toHaveAttribute('href', '/api/auth/login')
    expect(screen.getByRole('link', { name: 'Trámites' })).toHaveAttribute('aria-current', 'page')
    expect(screen.queryByRole('navigation', { name: 'Menú principal' })).not.toBeInTheDocument()
  })

  it('shows the home page with the account of the person with a session', async () => {
    stubApi({ session: true })
    const router = renderAt('/')

    expect(await screen.findByRole('heading', { level: 1, name: 'Mis trámites' })).toBeVisible()
    expect(screen.getByRole('button', { name: 'Cuenta de Ana Titular' })).toHaveTextContent('AT')
    expect(router.state.location.pathname).toBe('/')
  })

  it('shows the page about SIGTA to a visitor without a session', async () => {
    stubApi({ session: false })
    const router = renderAt('/about')

    expect(await screen.findByRole('heading', { level: 1, name: 'Acerca de SIGTA' })).toBeVisible()
    expect(router.state.location.pathname).toBe('/about')
  })

  it('sends someone with a session from the login page to the home page', async () => {
    stubApi({ session: true })
    const router = renderAt('/login')

    expect(await screen.findByRole('heading', { level: 1, name: 'Mis trámites' })).toBeVisible()
    expect(router.state.location.pathname).toBe('/')
  })
  it('opens the landing for a visitor when the API cannot say who is there', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ message: 'Bad gateway' }, { status: 502 })),
    )
    const router = renderAt('/login')

    expect(await screen.findByRole('link', { name: 'Iniciar sesión' })).toBeVisible()
    expect(router.state.location.pathname).toBe('/login')
  })
})
