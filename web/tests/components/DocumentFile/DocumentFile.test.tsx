import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { DocumentFile } from '../../../src/components/DocumentFile/DocumentFile'
import { render } from '../../test-utils/render'

const pdf = { id: 'D-1', file_name: 'constancia.pdf', media_type: 'application/pdf', size: 2048 }

function stubContent(status: number, type: string) {
  const fetch = vi.fn(
    async (_request: Request) => new Response('contenido', { status, headers: { 'Content-Type': type } }),
  )
  vi.stubGlobal('fetch', fetch)
  return fetch
}

describe('DocumentFile', () => {
  const revoke = vi.fn()

  beforeEach(() => {
    vi.stubGlobal('URL', Object.assign(URL, { createObjectURL: () => 'blob:preview', revokeObjectURL: revoke }))
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    revoke.mockClear()
  })

  it('names the file and its size and downloads it from the server', () => {
    render(<DocumentFile procedureId="P 1" file={pdf} />)

    expect(screen.getByText('constancia.pdf')).toBeDefined()
    expect(screen.getByText(/^2\sKB$/)).toBeDefined()
    const link = screen.getByRole('link', { name: 'Descargar constancia.pdf' })
    expect(link.getAttribute('href')).toBe('/api/procedures/P%201/documents/D-1/content')
    expect(link.getAttribute('download')).toBe('constancia.pdf')
  })

  it('shows in bytes a file smaller than one KB', () => {
    render(<DocumentFile procedureId="P-1" file={{ ...pdf, size: 512 }} />)

    expect(screen.getByText(/^512\sB$/)).toBeDefined()
  })

  it('names the procedure in its buttons, where files of several procedures share a page', () => {
    render(<DocumentFile procedureId="P-1" file={pdf} procedureName="Pase de carrera" />)

    expect(screen.getByRole('link', { name: 'Descargar constancia.pdf de Pase de carrera' })).toBeDefined()
    expect(screen.getByRole('button', { name: 'Ver constancia.pdf de Pase de carrera' })).toBeDefined()
  })

  it('shows a PDF in the page, from the content the server sent', async () => {
    const fetch = stubContent(200, 'application/pdf')
    render(<DocumentFile procedureId="P-1" file={pdf} />)

    await userEvent.click(screen.getByRole('button', { name: 'Ver constancia.pdf' }))

    const frame = await screen.findByTitle('constancia.pdf')
    expect(frame.tagName).toBe('IFRAME')
    expect(frame.getAttribute('src')).toBe('blob:preview')
    expect(fetch.mock.calls[0]![0].url).toMatch(/\/api\/procedures\/P-1\/documents\/D-1\/content$/)
  })

  it('shows an image in the page', async () => {
    stubContent(200, 'image/png')
    render(<DocumentFile procedureId="P-1" file={{ id: 'D-2', file_name: 'foto.png', media_type: 'image/png' }} />)

    await userEvent.click(screen.getByRole('button', { name: 'Ver foto.png' }))

    expect((await screen.findByRole('img', { name: 'foto.png' })).getAttribute('src')).toBe('blob:preview')
  })

  it('takes the type from the name when the file comes without it', () => {
    render(<DocumentFile procedureId="P-1" file={{ id: 'D-3', file_name: 'acta.PDF' }} />)

    expect(screen.getByRole('button', { name: 'Ver acta.PDF' })).toBeDefined()
  })

  it('only offers to download what the browser cannot show', () => {
    render(<DocumentFile procedureId="P-1" file={{ id: 'D-4', file_name: 'notas.txt', media_type: 'text/plain' }} />)

    expect(screen.queryByRole('button', { name: 'Ver notas.txt' })).toBeNull()
    expect(screen.getByRole('link', { name: 'Descargar notas.txt' })).toBeDefined()
  })

  it('only offers to download a PDF where the browser has no viewer', () => {
    vi.stubGlobal('navigator', { ...navigator, pdfViewerEnabled: false })
    render(<DocumentFile procedureId="P-1" file={pdf} />)

    expect(screen.queryByRole('button', { name: 'Ver constancia.pdf' })).toBeNull()
  })

  it('does not show a type that the server did not send as one it can show', async () => {
    stubContent(200, 'text/html')
    render(<DocumentFile procedureId="P-1" file={pdf} />)

    await userEvent.click(screen.getByRole('button', { name: 'Ver constancia.pdf' }))

    expect(
      await screen.findByText('Este archivo no se puede ver en el navegador. Descargalo para abrirlo.'),
    ).toBeDefined()
    expect(screen.queryByTitle('constancia.pdf')).toBeNull()
  })

  it('says when the file cannot be opened', async () => {
    stubContent(404, 'application/json')
    render(<DocumentFile procedureId="P-1" file={pdf} />)

    await userEvent.click(screen.getByRole('button', { name: 'Ver constancia.pdf' }))

    expect((await screen.findByRole('alert')).textContent).toBe(
      'No pudimos abrir el archivo. Volvé a intentarlo o descargalo.',
    )
  })

  it('frees the content when it is closed', async () => {
    stubContent(200, 'application/pdf')
    render(<DocumentFile procedureId="P-1" file={pdf} />)

    await userEvent.click(screen.getByRole('button', { name: 'Ver constancia.pdf' }))
    await screen.findByTitle('constancia.pdf')
    await userEvent.keyboard('{Escape}')

    await vi.waitFor(() => expect(revoke).toHaveBeenCalledWith('blob:preview'))
  })
})
