import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { CatalogPage } from '../../src/pages/CatalogPage'
import { render } from '../render'

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
      <CatalogPage />
    </QueryClientProvider>,
  )
}

describe('CatalogPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('shows what each procedure type asks for', async () => {
    renderCatalog(200, [equivalence])

    expect(await screen.findByRole('heading', { name: 'Trámite B' })).toBeInTheDocument()
    expect(screen.getByText('Una descripción llana')).toBeInTheDocument()
    expect(screen.getByText('Nombre (Obligatorio)')).toBeInTheDocument()
    expect(screen.getByText('Motivo (Opcional)')).toBeInTheDocument()
    expect(screen.getByText('Carta (Obligatorio). Formatos: PDF, WEBP. Tamaño máximo: 10 MB.')).toBeInTheDocument()
    expect(screen.getByText('Hasta 30 días antes')).toBeInTheDocument()
    expect(screen.getByText('Ser regular')).toBeInTheDocument()
  })

  it('sorts the procedure types by name', async () => {
    renderCatalog(200, [equivalence, { ...equivalence, id: 2, name: 'Ámbito' }])

    await screen.findByRole('heading', { name: 'Trámite B' })
    const names = screen.getAllByRole('heading', { level: 2 }).map((heading) => heading.textContent)
    expect(names).toEqual(['Ámbito', 'Trámite B'])
  })

  it('leaves out the sections a procedure type does not have', async () => {
    renderCatalog(200, [{ ...equivalence, deadlines: [], admissibility: [] }])

    await screen.findByRole('heading', { name: 'Trámite B' })
    expect(screen.queryByRole('heading', { name: 'Plazos' })).not.toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Condiciones de admisibilidad' })).not.toBeInTheDocument()
  })

  it('shows a document of an older format without types or size', async () => {
    const document = { id: 'letter', name: 'Carta', required: true, accepted_types: [], max_size: 0 }
    renderCatalog(200, [{ ...equivalence, description: '', documents: [document] }])

    expect(await screen.findByText('Carta (Obligatorio).')).toBeInTheDocument()
  })

  it('shows repeated deadlines', async () => {
    renderCatalog(200, [{ ...equivalence, deadlines: ['Hasta el viernes', 'Hasta el viernes'] }])

    expect(await screen.findAllByText('Hasta el viernes')).toHaveLength(2)
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
