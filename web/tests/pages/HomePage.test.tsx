import { screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
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
})
