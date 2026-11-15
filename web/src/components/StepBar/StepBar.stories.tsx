import { Stack } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { StepBar } from './StepBar'

const meta: Meta = {
  title: 'SIGTA/StepBar',
  parameters: {
    docs: {
      description: {
        component:
          'La etapa de un trámite entre los pasos de su tipo: un segmento por paso, de 4 de alto, 4 entre segmentos y radio completo. Los hechos en `action-600`, el abierto en `brand-500` y los que faltan en `border-subtle`. Un lector de pantalla lee "Etapa n de m".',
      },
    },
  },
}

export default meta

export const Steps: StoryObj = {
  name: 'StepBar',
  render: () => (
    <Stack gap="md" w="var(--sigta-measure)">
      <StepBar step={1} count={9} />
      <StepBar step={4} count={9} />
      <StepBar step={9} count={9} />
    </Stack>
  ),
}
