import type { Meta, StoryObj } from '@storybook/react-vite'
import { applicantHowTo } from '../../content/platform'
import { StepList } from './StepList'

const meta: Meta<typeof StepList> = {
  title: 'SIGTA/StepList',
  component: StepList,
  parameters: {
    docs: {
      description: {
        component:
          'Cómo se hace algo, en orden: una lista ordenada en tres columnas, una en el celular, sin cajas. El número va en el color de acción, sin círculo, y una línea `border-subtle` arriba de cada columna marca la secuencia.',
      },
    },
  },
}

export default meta

export const Steps: StoryObj<typeof StepList> = {
  name: 'StepList',
  args: { steps: applicantHowTo },
}
