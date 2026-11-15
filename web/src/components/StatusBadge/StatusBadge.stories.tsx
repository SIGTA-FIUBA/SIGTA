import { Group } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { StatusBadge } from './StatusBadge'

const meta: Meta<typeof StatusBadge> = {
  title: 'SIGTA/StatusBadge',
  component: StatusBadge,
  parameters: {
    docs: {
      description: {
        component:
          'El estado de un trámite: alto 24, radio completo, `caption` en negrita, el tono 700 sobre el tinte 50 de su color. El estado se lee en el texto, no solo en el color. Los nombres de los estados los define la Secretaría.',
      },
    },
  },
}

export default meta

export const Tones: StoryObj<typeof StatusBadge> = {
  name: 'StatusBadge',
  render: () => (
    <Group gap="sm">
      <StatusBadge tone="office">En la Secretaría</StatusBadge>
      <StatusBadge tone="external">En evaluación externa</StatusBadge>
      <StatusBadge tone="applicant">Pendiente del titular</StatusBadge>
      <StatusBadge tone="done">Resuelto</StatusBadge>
      <StatusBadge tone="error">Rechazado</StatusBadge>
    </Group>
  ),
}
