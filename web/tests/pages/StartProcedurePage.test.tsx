import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { StartProcedurePage } from '../../src/pages/StartProcedurePage'
import { render } from '../test-utils/render'

const pdf = (name: string, size = 200) => new File(['%PDF-1.7\n'.padEnd(size, 'x')], name, { type: 'application/pdf' })

const transcript = {
  id: 'transcript',
  name: 'Historia académica',
  required: true,
  accepted_types: ['application/pdf'],
  max_size: 1024,
}
const program = {
  id: 'program',
  name: 'Programa',
  required: false,
  accepted_types: ['application/pdf'],
  max_size: 1024,
}

const entry = {
  id: 7,
  version: 1,
  name: 'Equivalencia de prueba',
  description: '',
  deadlines: [],
  admissibility: [],
  fields: [],
  documents: [],
  submission: {
    fields: [
      { id: 'reason', name: 'Motivo', required: true, type: 'long_text', options: [] },
      { id: 'credits', name: 'Créditos', required: true, type: 'number', options: [] },
      {
        id: 'modality',
        name: 'Modalidad',
        required: true,
        type: 'option',
        options: [
          { id: 'full', name: 'Total' },
          { id: 'partial', name: 'Parcial' },
        ],
      },
      { id: 'subjects', name: 'Asignaturas', required: false, type: 'subject_list', options: [] },
    ],
    documents: [transcript, program],
  },
}

type Answer = { status: number; body: unknown }

// stubApi answers the catalog and records the body of each submission it
// receives, as text: undici cannot parse a form with the File of jsdom.
function stubApi(answer: Answer = { status: 201, body: { id: 'P-42', state: 'in_office', open_step: 'review' } }) {
  const sent: string[] = []
  vi.stubGlobal(
    'fetch',
    vi.fn(async (request: Request) => {
      if (request.method === 'POST') {
        sent.push(await request.text())
        return Response.json(answer.body, { status: answer.status })
      }
      return Response.json([entry])
    }),
  )
  return sent
}

const lineBreak = '\r\n'

// parse reads the name, the file name and the value of each part of a
// multipart body, in order.
function parse(body: string) {
  const boundary = body.slice(0, body.indexOf(lineBreak))
  return body
    .split(boundary)
    .slice(1, -1)
    .map((part) => {
      const [head, ...rest] = part.split(lineBreak + lineBreak)
      const value = rest.join(lineBreak + lineBreak)
      return {
        name: /name="([^"]*)"/.exec(head)?.[1],
        fileName: /filename="([^"]*)"/.exec(head)?.[1],
        value: value.endsWith(lineBreak) ? value.slice(0, -lineBreak.length) : value,
      }
    })
}

function renderAt(path: string) {
  const router = createMemoryRouter([{ path: '/catalog/:procedureTypeId/start', element: <StartProcedurePage /> }], {
    initialEntries: [path],
  })
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

function fileInputs() {
  return Array.from(document.querySelectorAll<HTMLInputElement>('input[type="file"]'))
}

async function fillData(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText(/Motivo/), 'Cursé la materia afuera')
  await user.type(screen.getByLabelText(/Créditos/), '8,5')
  await user.click(screen.getByRole('combobox', { name: 'Modalidad' }))
  await user.click(screen.getByRole('option', { name: 'Parcial' }))
  await user.type(screen.getByRole('combobox', { name: 'Asignaturas' }), 'Física I{Enter}Química{Enter}')
  await user.click(screen.getByRole('button', { name: 'Continuar' }))
}

async function reachSummary(user: ReturnType<typeof userEvent.setup>) {
  await screen.findByLabelText(/Motivo/)
  await fillData(user)
  await user.upload(fileInputs()[0], pdf('historia.pdf'))
  await user.click(screen.getByRole('button', { name: 'Ver el resumen' }))
}

describe('StartProcedurePage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('builds the form from what the procedure type asks for', async () => {
    stubApi()
    renderAt('/catalog/7/start')

    expect(await screen.findByRole('heading', { level: 1, name: 'Equivalencia de prueba' })).toBeVisible()
    expect(screen.getByLabelText(/Motivo/)).toBeVisible()
    expect(screen.getByLabelText(/Créditos/)).toBeVisible()
    expect(screen.getByRole('combobox', { name: 'Modalidad' })).toBeVisible()
    expect(screen.getByRole('combobox', { name: 'Asignaturas' })).toBeVisible()
  })

  it('does not move on while a mandatory field is empty', async () => {
    stubApi()
    const user = userEvent.setup()
    renderAt('/catalog/7/start')

    await user.click(await screen.findByRole('button', { name: 'Continuar' }))

    expect(screen.getAllByText('Completá este dato.')).toHaveLength(3)
    expect(screen.queryByText(/Historia académica/)).not.toBeInTheDocument()
  })

  it('checks the format of a field before moving on', async () => {
    stubApi()
    const user = userEvent.setup()
    renderAt('/catalog/7/start')

    await user.type(await screen.findByLabelText(/Créditos/), 'doce')
    await user.click(screen.getByRole('button', { name: 'Continuar' }))

    expect(screen.getByText('Escribí un número, con coma o punto para los decimales.')).toBeVisible()
  })

  it('takes the focus to the first field to fix', async () => {
    stubApi()
    const user = userEvent.setup()
    renderAt('/catalog/7/start')

    await user.type(await screen.findByLabelText(/Créditos/), 'doce')
    await user.click(screen.getByRole('button', { name: 'Continuar' }))

    expect(screen.getByLabelText(/Motivo/)).toHaveFocus()
  })

  it('asks before leaving with something filled in', async () => {
    stubApi()
    const user = userEvent.setup()
    renderAt('/catalog/7/start')

    await user.type(await screen.findByLabelText(/Motivo/), 'Cursé la materia afuera')
    await user.click(screen.getByRole('link', { name: 'Cancelar' }))

    expect(await screen.findByRole('dialog', { name: '¿Salir sin enviar el trámite?' })).toBeVisible()
    await user.click(screen.getByRole('button', { name: 'Seguir completando' }))
    expect(screen.getByLabelText(/Motivo/)).toHaveValue('Cursé la materia afuera')
  })

  it('does not move on without a mandatory document', async () => {
    stubApi()
    const user = userEvent.setup()
    renderAt('/catalog/7/start')
    await screen.findByLabelText(/Motivo/)

    await fillData(user)
    await user.click(screen.getByRole('button', { name: 'Ver el resumen' }))

    expect(screen.getByText('Adjuntá este documento.')).toBeVisible()
  })

  it('rejects a file larger than the document allows', async () => {
    stubApi()
    const user = userEvent.setup()
    renderAt('/catalog/7/start')
    await screen.findByLabelText(/Motivo/)
    await fillData(user)

    await user.upload(fileInputs()[0], pdf('grande.pdf', 2048))
    await user.click(screen.getByRole('button', { name: 'Ver el resumen' }))

    expect(screen.getByText('El archivo supera el tamaño máximo de 1 KB.')).toBeVisible()
  })

  it('shows a summary, sends everything at once and shows the identifier', async () => {
    const sent = stubApi()
    const user = userEvent.setup()
    renderAt('/catalog/7/start')
    await screen.findByLabelText(/Motivo/)
    await fillData(user)

    await user.upload(fileInputs()[0], pdf('historia.pdf'))
    await user.upload(fileInputs()[0], pdf('historia-corregida.pdf'))
    await user.upload(fileInputs()[1], pdf('programa.pdf'))
    await user.click(screen.getByRole('button', { name: 'Quitar el archivo de Programa' }))
    await user.click(screen.getByRole('button', { name: 'Ver el resumen' }))

    const value = (name: string) => screen.getByText(name).nextElementSibling
    expect(value('Motivo')).toHaveTextContent('Cursé la materia afuera')
    expect(value('Créditos')).toHaveTextContent('8,5')
    expect(value('Modalidad')).toHaveTextContent('Parcial')
    expect(value('Asignaturas')).toHaveTextContent('Física I, Química')
    expect(value('Historia académica')).toHaveTextContent('historia-corregida.pdf')
    expect(value('Programa')).toHaveTextContent('Sin adjuntar')
    expect(sent).toHaveLength(0)

    await user.click(screen.getByRole('button', { name: 'Enviar el trámite' }))

    expect(await screen.findByText('P-42')).toBeVisible()
    expect(screen.getByRole('link', { name: 'Ver el trámite' })).toHaveAttribute('href', '/procedures/P-42')
    expect(sent).toHaveLength(1)
    const parts = parse(sent[0])
    expect(parts.map((part) => part.name)).toEqual(['procedure_type', 'values', 'document', 'file'])
    expect(parts[0].value).toBe('7')
    expect(JSON.parse(parts[1].value)).toEqual({
      reason: 'Cursé la materia afuera',
      credits: '8.5',
      modality: 'partial',
      subjects: ['Física I', 'Química'],
    })
    expect(parts[2].value).toBe('transcript')
    // jsdom hands the file to the fetch of Node as an empty Blob, without its
    // name or content: the summary above shows which file replaced the first.
  })

  it('takes the applicant back to what the server rejected', async () => {
    stubApi({
      status: 400,
      body: {
        message: 'El trámite no cumple con lo que pide.',
        problems: [{ code: 'type_not_accepted', kind: 'document', id: 'transcript' }],
      },
    })
    const user = userEvent.setup()
    renderAt('/catalog/7/start')
    await reachSummary(user)

    await user.click(screen.getByRole('button', { name: 'Enviar el trámite' }))

    expect(await screen.findByText('El archivo no es de un formato que acepte este documento.')).toBeVisible()
    expect(screen.getByRole('button', { name: 'Ver el resumen' })).toBeVisible()
  })

  it('says what to do when the submission fails', async () => {
    stubApi({ status: 500, body: { message: 'Ocurrió un error inesperado.' } })
    const user = userEvent.setup()
    renderAt('/catalog/7/start')
    await reachSummary(user)

    await user.click(screen.getByRole('button', { name: 'Enviar el trámite' }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'No pudimos enviar el trámite. Volvé a intentarlo más tarde.',
    )
  })

  it('says when the procedure type does not exist', async () => {
    stubApi()
    renderAt('/catalog/99/start')

    expect(await screen.findByText(/No existe ese trámite/)).toBeVisible()
  })

  it('says when the catalog fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ message: 'Ocurrió un error inesperado.' }, { status: 500 })),
    )
    renderAt('/catalog/7/start')

    expect(await screen.findByRole('alert')).toHaveTextContent('No pudimos cargar el trámite')
  })
})
