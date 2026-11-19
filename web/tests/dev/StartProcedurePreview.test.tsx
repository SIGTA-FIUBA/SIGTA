import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { describe, expect, it } from 'vitest'
import { StartProcedurePreview } from '../../src/dev/StartProcedurePreview'
import { fixtures } from '../../src/dev/startProcedureFixtures'
import { render } from '../test-utils/render'

function renderPreview(data: string) {
  const router = createMemoryRouter(
    [{ path: '/dev/start-procedure/:procedureTypeId', element: <StartProcedurePreview /> }],
    { initialEntries: [`/dev/start-procedure/1?data=${data}`] },
  )
  render(<RouterProvider router={router} />)
}

describe('StartProcedurePreview', () => {
  it.each(Object.entries(fixtures))('renders the %s catalog without asking the API', async (data, entry) => {
    renderPreview(data)

    expect(await screen.findByRole('heading', { level: 1, name: entry.name })).toBeVisible()
  })

  it('says what the summary has when the procedure type asks for nothing', async () => {
    const user = userEvent.setup()
    renderPreview('empty')

    await user.click(await screen.findByRole('button', { name: 'Continuar' }))
    await user.click(screen.getByRole('button', { name: 'Ver el resumen' }))

    expect(screen.getByText('Este trámite no pide datos ni documentos.')).toBeVisible()
  })
})
