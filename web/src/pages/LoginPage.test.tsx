import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { describe, expect, it } from 'vitest'
import { LoginPage } from './LoginPage'

function renderAt(path: string) {
  const router = createMemoryRouter([{ path: '/login', element: <LoginPage /> }], {
    initialEntries: [path],
  })
  render(<RouterProvider router={router} />)
}

describe('LoginPage', () => {
  it('leaves for the identity provider through the API', () => {
    renderAt('/login')

    expect(screen.getByRole('link', { name: 'Iniciar sesión con Google' })).toHaveAttribute('href', '/api/auth/login')
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
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
