import { screen } from '@testing-library/react'
import { createMemoryRouter, MemoryRouter, Outlet, RouterProvider } from 'react-router'
import { describe, expect, it } from 'vitest'
import { render } from '../render'
import { HomePage } from '../../src/pages/HomePage'

describe('HomePage', () => {
  it('shows the title of the page', () => {
    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { name: 'Inicio' })).toBeVisible()
  })

  it('links to the catalog', () => {
    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>,
    )

    expect(screen.getByRole('link', { name: 'Ver los trámites disponibles' })).toHaveAttribute('href', '/catalog')
  })

  it('links to the inbox only for the staff', () => {
    const page = (role: string) =>
      createMemoryRouter([
        {
          element: <Outlet context={{ id: 'u1', email: 'u@fi.uba.ar', name: 'Persona', role }} />,
          children: [{ path: '/', element: <HomePage /> }],
        },
      ])

    const { unmount } = render(<RouterProvider router={page('staff_operative')} />)
    expect(screen.getByRole('link', { name: 'Ver la bandeja' })).toHaveAttribute('href', '/inbox')
    unmount()

    render(<RouterProvider router={page('applicant')} />)
    expect(screen.queryByRole('link', { name: 'Ver la bandeja' })).not.toBeInTheDocument()
  })
})
