import { screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { routes } from './routes'
import { render } from './test/render'

const user = { id: 'u1', email: 'titular@fi.uba.ar', name: 'Titular', role: 'applicant' }

function withSession(status: number) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => Response.json(status === 200 ? user : { message: 'Sin sesión' }, { status })),
  )
}

function renderAt(path: string) {
  const router = createMemoryRouter(routes, { initialEntries: [path] })
  render(<RouterProvider router={router} />)
  return router
}

describe('routes', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('sends a visitor without a session to the login page', async () => {
    withSession(401)
    const router = renderAt('/')

    expect(await screen.findByRole('link', { name: 'Iniciar sesión con Google' })).toBeVisible()
    expect(router.state.location.pathname).toBe('/login')
  })

  it('shows the home page with a session', async () => {
    withSession(200)
    const router = renderAt('/')

    expect(await screen.findByRole('heading', { name: 'SIGTA' })).toBeVisible()
    expect(router.state.location.pathname).toBe('/')
  })

  it('sends someone with a session from the login page to the home page', async () => {
    withSession(200)
    const router = renderAt('/login')

    expect(await screen.findByRole('heading', { name: 'SIGTA' })).toBeVisible()
    expect(router.state.location.pathname).toBe('/')
  })
})
