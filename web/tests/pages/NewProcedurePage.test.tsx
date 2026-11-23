import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { NewProcedurePage } from '../../src/pages/NewProcedurePage'
import { render } from '../test-utils/render'

const procedureType = (id: number, name: string) => ({
  id,
  version: 1,
  name,
  description: 'Una descripción',
  deadlines: [],
  admissibility: [],
  fields: [{ id: 'name', name: 'Nombre', required: true }],
  documents: [],
})

function renderPage(status: number, body: unknown) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => Response.json(body, { status })),
  )
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <NewProcedurePage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('NewProcedurePage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('offers every procedure type to start, by name, without what it asks for', async () => {
    renderPage(200, [procedureType(1, 'Trámite B'), procedureType(2, 'Ámbito')])

    await screen.findByRole('link', { name: /Trámite B/ })
    const links = screen.getAllByRole('link', { name: /Una descripción/ })
    expect(links.map((link) => link.getAttribute('href'))).toEqual(['/catalog/2/start', '/catalog/1/start'])
    expect(screen.queryByText(/dato/)).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'catálogo de trámites' })).toHaveAttribute('href', '/catalog')
  })

  it('says when the catalog fails', async () => {
    renderPage(500, { message: 'Ocurrió un error inesperado.' })

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudo cargar el catálogo')
  })
})
