import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { CatalogPage } from '../../src/pages/CatalogPage'
import { render } from '../test-utils/render'

const equivalence = {
  id: 1,
  version: 2,
  name: 'Trámite B',
  description: 'Una descripción llana',
  deadlines: ['Hasta 30 días antes'],
  admissibility: ['Ser regular'],
  fields: [
    { id: 'name', name: 'Nombre', required: true },
    { id: 'reason', name: 'Motivo', required: false },
  ],
  documents: [
    {
      id: 'letter',
      name: 'Carta',
      required: true,
      accepted_types: ['application/pdf', 'image/webp'],
      max_size: 10485760,
    },
  ],
}

function renderCatalog(status: number, body: unknown) {
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => Response.json(body, { status })),
  )
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <CatalogPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('CatalogPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('shows what each procedure type is, who can ask for it and its deadlines', async () => {
    renderCatalog(200, [equivalence])

    await userEvent.click(await screen.findByRole('button', { name: /Trámite B/ }))

    expect(screen.getByText('Una descripción llana')).toBeVisible()
    expect(screen.getByText('Ser regular')).toBeVisible()
    expect(screen.getByText('Hasta 30 días antes')).toBeVisible()
  })

  it('does not show the fields, documents, formats or sizes the form asks for', async () => {
    renderCatalog(200, [equivalence])

    await userEvent.click(await screen.findByRole('button', { name: /Trámite B/ }))

    for (const text of ['Nombre', 'Motivo', 'Carta', 'Obligatorio', 'PDF', '10 MB']) {
      expect(screen.queryByText(new RegExp(text))).not.toBeInTheDocument()
    }
  })

  it('links each procedure type to the form that starts it', async () => {
    renderCatalog(200, [equivalence])

    await userEvent.click(await screen.findByRole('button', { name: /Trámite B/ }))

    expect(await screen.findByRole('link', { name: 'Iniciar Trámite B' })).toHaveAttribute('href', '/catalog/1/start')
  })

  it('sorts the procedure types by name', async () => {
    renderCatalog(200, [equivalence, { ...equivalence, id: 2, name: 'Ámbito' }])

    await screen.findByRole('heading', { name: 'Trámite B' })
    const names = screen.getAllByRole('heading', { level: 2 }).map((heading) => heading.textContent)
    expect(names).toEqual(['Ámbito', 'Trámite B'])
  })

  it('leaves out the sections a procedure type does not have', async () => {
    renderCatalog(200, [{ ...equivalence, deadlines: [], admissibility: [] }])

    await userEvent.click(await screen.findByRole('button', { name: /Trámite B/ }))

    expect(screen.queryByText('Plazos')).not.toBeInTheDocument()
    expect(screen.queryByText('Quién puede pedirlo')).not.toBeInTheDocument()
  })

  it('shows repeated deadlines', async () => {
    renderCatalog(200, [{ ...equivalence, deadlines: ['Hasta el viernes', 'Hasta el viernes'] }])

    expect(await screen.findAllByText('Hasta el viernes', {}, { timeout: 2000 })).toHaveLength(2)
  })

  it('says when there are no procedure types', async () => {
    renderCatalog(200, [])

    expect(await screen.findByText('Todavía no hay trámites disponibles.')).toBeInTheDocument()
  })

  it('says when the catalog fails', async () => {
    renderCatalog(500, { message: 'Ocurrió un error inesperado.' })

    expect(await screen.findByRole('alert')).toHaveTextContent('No pudimos cargar los trámites')
  })
})
