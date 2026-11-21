import { Stack } from '@mantine/core'
import { useForm } from '@mantine/form'
import type { Meta, StoryObj } from '@storybook/react-vite'
import type { components } from '../../api/schema'
import type { SubmissionDraft } from '../../submission'
import { FieldInput } from './FieldInput'

type FieldRequirement = components['schemas']['FieldRequirement']

const meta: Meta = {
  title: 'SIGTA/FieldInput',
  parameters: {
    docs: {
      description: {
        component:
          'El campo de un tipo de trámite, según su tipo, al iniciar un trámite. Los campos obligatorios llevan asterisco y el error aparece debajo.',
      },
    },
  },
}

export default meta

const fields: FieldRequirement[] = [
  { id: 'reason', name: 'Motivo', required: true, type: 'text', options: [] },
  { id: 'detail', name: 'Detalle', required: false, type: 'long_text', options: [] },
  { id: 'credits', name: 'Créditos', required: true, type: 'number', options: [] },
  { id: 'since', name: 'Fecha de aprobación', required: true, type: 'date', options: [] },
  {
    id: 'term',
    name: 'Cuatrimestre',
    required: true,
    type: 'option',
    options: [
      { id: 'first', name: 'Primer cuatrimestre' },
      { id: 'second', name: 'Segundo cuatrimestre' },
    ],
  },
  { id: 'subjects', name: 'Asignaturas', required: true, type: 'subject_list', options: [] },
]

function Fields({ error }: { error?: boolean }) {
  const form = useForm<SubmissionDraft>({
    mode: 'uncontrolled',
    initialValues: { values: { subjects: ['Análisis Matemático II'] }, files: {} },
    initialErrors: error ? { 'values.reason': 'Completá este campo.' } : {},
  })
  return (
    <Stack maw="var(--sigta-measure)">
      {fields.map((field) => (
        <FieldInput key={field.id} form={form} field={field} />
      ))}
    </Stack>
  )
}

export const Types: StoryObj = {
  name: 'Por tipo de campo',
  render: () => <Fields />,
}

export const WithError: StoryObj = {
  name: 'Con error',
  render: () => <Fields error />,
}
