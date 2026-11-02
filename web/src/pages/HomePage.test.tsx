import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { describe, expect, it } from 'vitest'
import { HomePage } from './HomePage'

describe('HomePage', () => {
  it('shows the name of the platform', () => {
    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>,
    )

    expect(screen.getByRole('heading', { name: 'SIGTA' })).toBeInTheDocument()
  })
})
