import { screen } from '@testing-library/react'
import { createMemoryRouter, MemoryRouter, Outlet, RouterProvider } from 'react-router'
import { describe, expect, it } from 'vitest'
import { render } from '../render'
import { HomePage } from '../../src/pages/HomePage'

function renderAs(role: string) {
  const router = createMemoryRouter([
    {
      element: <Outlet context={{ id: 'u1', email: 'u@fi.uba.ar', name: 'Persona', role }} />,
      children: [{ path: '/', element: <HomePage /> }],
    },
  ])
  render(<RouterProvider router={router} />)
}

describe('HomePage', () => {
  it('says what the platform is', () => {
    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>,
    )

    expect(
      screen.getByRole('heading', { level: 1, name: 'Sistema Integral de Gestión de Trámites Académicos' }),
    ).toBeVisible()
    expect(screen.getByRole('heading', { name: 'Criterios de la plataforma' })).toBeVisible()
  })

  it('explains to an applicant how a procedure is done, and links to the catalog', () => {
    renderAs('applicant')

    expect(screen.getByRole('heading', { name: 'Cómo se realiza un trámite' })).toBeVisible()
    expect(screen.getByRole('link', { name: /Catálogo de trámites/ })).toHaveAttribute('href', '/catalog')
    expect(screen.queryByRole('link', { name: /Bandeja de trámites/ })).not.toBeInTheDocument()
  })

  it('explains to the staff how to work, and links to the inbox', () => {
    renderAs('staff_operative')

    expect(screen.getByRole('heading', { name: 'Cómo se trabaja en SIGTA' })).toBeVisible()
    expect(screen.getByRole('link', { name: /Bandeja de trámites/ })).toHaveAttribute('href', '/inbox')
  })
})
