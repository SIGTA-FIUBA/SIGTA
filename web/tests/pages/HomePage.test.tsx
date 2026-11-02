import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { render } from '../render'
import { HomePage } from '../../src/pages/HomePage'

describe('HomePage', () => {
  it('shows the title of the page', () => {
    render(<HomePage />)

    expect(screen.getByRole('heading', { name: 'Inicio' })).toBeVisible()
  })
})
