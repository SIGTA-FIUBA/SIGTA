import { screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { describe, expect, it } from 'vitest'
import { AdministrationPreview } from '../../src/dev/AdministrationPreview'
import { fixtures } from '../../src/dev/administrationFixtures'
import { render } from '../test-utils/render'

function renderPreview(data: string) {
  const router = createMemoryRouter([{ path: '/dev/administration', element: <AdministrationPreview /> }], {
    initialEntries: [`/dev/administration?data=${data}`],
  })
  render(<RouterProvider router={router} />)
}

describe('AdministrationPreview', () => {
  it.each(Object.entries(fixtures))('renders the %s areas without asking the API', async (data, fixture) => {
    renderPreview(data)

    for (const area of fixture.areas) {
      expect(await screen.findByRole('region', { name: area.name })).toBeVisible()
    }
  })
})
