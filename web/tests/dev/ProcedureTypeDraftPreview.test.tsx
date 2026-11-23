import { screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { describe, expect, it } from 'vitest'
import { ProcedureTypeDraftPreview } from '../../src/dev/ProcedureTypeDraftPreview'
import { fixtures } from '../../src/dev/procedureTypeDraftFixtures'
import { render } from '../test-utils/render'

function renderPreview(data: string) {
  const router = createMemoryRouter(
    [{ path: '/dev/procedure-types/:procedureTypeId/draft', element: <ProcedureTypeDraftPreview /> }],
    { initialEntries: [`/dev/procedure-types/1/draft?data=${data}`] },
  )
  render(<RouterProvider router={router} />)
}

describe('ProcedureTypeDraftPreview', () => {
  it.each(Object.entries(fixtures))('renders the %s draft without asking the API', async (data, fixture) => {
    renderPreview(data)

    expect(await screen.findByRole('heading', { level: 1, name: fixture.draft.definition.name })).toBeVisible()
    for (const [index, step] of fixture.draft.definition.steps.entries()) {
      expect(screen.getByRole('button', { name: new RegExp(`^Paso ${index + 1}: `) })).toHaveTextContent(step.name)
    }
  })
})
