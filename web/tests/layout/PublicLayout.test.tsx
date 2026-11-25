import { fireEvent, screen, within } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { describe, expect, it, vi } from 'vitest'
import { PublicLayout } from '../../src/layout/PublicLayout'
import { render } from '../test-utils/render'

function renderAt(path: string) {
  const router = createMemoryRouter(
    [
      {
        element: <PublicLayout />,
        children: ['/login', '/catalog', '/about'].map((page) => ({ path: page, element: <p>{page}</p> })),
      },
    ],
    { initialEntries: [path] },
  )
  render(<RouterProvider router={router} />)
}

describe('PublicLayout', () => {
  it('leaves for the identity provider through the API', () => {
    renderAt('/login')

    expect(screen.getByRole('link', { name: 'Iniciar sesión' })).toHaveAttribute('href', '/api/auth/login')
    expect(screen.getByRole('link', { name: 'SIGTA, inicio' })).toHaveAttribute('href', '/login')
  })

  it.each([
    ['/login', 'Inicio'],
    ['/catalog', 'Trámites'],
    ['/about', 'Sobre SIGTA'],
  ])('leads to the public pages from its header, on %s', (path, current) => {
    renderAt(path)

    const nav = screen.getByRole('navigation', { name: 'SIGTA' })
    expect(
      within(nav)
        .getAllByRole('link')
        .map((link) => link.getAttribute('href')),
    ).toEqual(['/login', '/catalog', '/about'])
    expect(within(nav).getByRole('link', { current: 'page' })).toHaveTextContent(current)
    expect(screen.getByRole('main')).toHaveTextContent(path)
  })

  it('takes to the top of the page a link to the page the person is on', () => {
    renderAt('/login')
    const scroll = vi.spyOn(window, 'scrollTo').mockImplementation(() => {})

    fireEvent.click(screen.getByRole('link', { name: 'SIGTA, inicio' }))
    fireEvent.click(within(screen.getByRole('navigation', { name: 'SIGTA' })).getByRole('link', { name: 'Inicio' }))

    expect(scroll).toHaveBeenCalledTimes(2)
    expect(screen.getByRole('main')).toHaveTextContent('/login')
    scroll.mockRestore()
  })

  it('signs with SIGTA in the footer, a link to the home page', () => {
    renderAt('/catalog')

    expect(
      within(screen.getByRole('contentinfo')).getByRole('link', {
        name: 'SIGTA, Sistema Integral de Gestión de Trámites Académicos',
      }),
    ).toHaveAttribute('href', '/login')
  })
})
