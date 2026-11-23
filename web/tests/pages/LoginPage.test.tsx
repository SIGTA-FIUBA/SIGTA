import { fireEvent, screen, within } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { describe, expect, it, vi } from 'vitest'
import { render } from '../test-utils/render'
import { LoginPage } from '../../src/pages/LoginPage'

function renderAt(path: string) {
  const router = createMemoryRouter([{ path: '/login', element: <LoginPage /> }], {
    initialEntries: [path],
  })
  return render(<RouterProvider router={router} />)
}

describe('LoginPage', () => {
  it('asks for nothing: the login is the one of the identity provider', () => {
    renderAt('/login')

    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('opens with the headline and leads to what SIGTA offers', () => {
    renderAt('/login')

    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Una nueva forma de hacer trámites')
    expect(screen.getByRole('link', { name: 'Conocé SIGTA' })).toHaveAttribute('href', '#que-podes-hacer')
    expect(screen.getByRole('heading', { level: 2, name: '¿Qué podés hacer con SIGTA?' })).toHaveAttribute(
      'id',
      'que-podes-hacer',
    )
  })

  it('offers the inbox, the catalog and the administration, each with its way in', () => {
    renderAt('/login')

    expect(screen.getByRole('link', { name: 'Bandeja de trámites' })).toHaveAttribute('href', '/api/auth/login')
    expect(screen.getByRole('link', { name: 'Catálogo de trámites' })).toHaveAttribute('href', '/catalog')
    expect(screen.getByRole('link', { name: 'Administración de SIGTA' })).toHaveAttribute('href', '/about')
  })

  it('glides to what SIGTA offers from "Conocé SIGTA" and from its band', () => {
    renderAt('/login')
    const section = screen.getByRole('region', { name: '¿Qué podés hacer con SIGTA?' })
    const scroll = vi.spyOn(section, 'scrollIntoView')

    fireEvent.click(screen.getByRole('link', { name: 'Conocé SIGTA' }))
    fireEvent.click(screen.getByRole('link', { name: 'Ver qué podés hacer con SIGTA' }))

    expect(scroll).toHaveBeenCalledTimes(2)
  })

  it('makes the round shapes of the collage links for the pointer, out of the tab order', () => {
    const { container } = renderAt('/login')

    const shapes = [...container.querySelectorAll('[aria-hidden] a')]
    expect(shapes.map((shape) => shape.getAttribute('href'))).toEqual(['/about', '/catalog', '/api/auth/login'])
    expect(shapes.every((shape) => shape.getAttribute('tabindex') === '-1')).toBe(true)
  })

  it('says how SIGTA is used and the criteria it follows, in order', () => {
    renderAt('/login')

    const steps = screen.getByRole('heading', { name: '¿Cómo se trabaja en SIGTA?' }).closest('section')
    expect(
      within(steps as HTMLElement)
        .getAllByRole('heading', { level: 3 })
        .map((heading) => heading.textContent),
    ).toEqual(['Bandeja de trámites', 'Registro de lo actuado', 'Seguimiento de derivaciones'])
    const criteria = screen.getByRole('heading', { name: 'Criterios de la plataforma' }).closest('section')
    expect(within(criteria as HTMLElement).getAllByRole('listitem')).toHaveLength(3)
  })

  it.each([
    ['state_mismatch', 'El ingreso no coincide con el que se inició. Volvé a intentarlo.'],
    ['rejected', 'El ingreso no se pudo completar. Volvé a intentarlo.'],
    ['domain_not_allowed', 'Ingresá con tu cuenta institucional.'],
  ])('explains the %s error', (error, message) => {
    renderAt(`/login?error=${error}`)

    expect(screen.getByRole('alert')).toHaveTextContent(message)
  })

  it('treats an unknown error as a rejected login', () => {
    renderAt('/login?error=constructor')

    expect(screen.getByRole('alert')).toHaveTextContent('El ingreso no se pudo completar.')
  })
})
