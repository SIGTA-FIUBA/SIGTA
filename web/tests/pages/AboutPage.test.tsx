import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { render } from '../test-utils/render'
import { AboutPage } from '../../src/pages/AboutPage'

describe('AboutPage', () => {
  it('says what SIGTA is and the criteria it follows', () => {
    render(<AboutPage />)

    expect(screen.getByRole('heading', { level: 1, name: 'Acerca de SIGTA' })).toBeVisible()
    expect(screen.getByText(/Sistema Integral de Gestión de Trámites Académicos/)).toBeVisible()
    expect(screen.getByRole('heading', { name: 'Criterios de la plataforma' })).toBeVisible()
  })

  it('signs with FIUBA', () => {
    render(<AboutPage />)

    expect(screen.getByRole('img', { name: 'SIGTA' })).toBeVisible()
    expect(screen.getByRole('img', { name: 'Facultad de Ingeniería, Universidad de Buenos Aires' })).toBeVisible()
  })
})
