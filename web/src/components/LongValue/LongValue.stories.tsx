import type { Meta, StoryObj } from '@storybook/react-vite'
import { LongValue } from './LongValue'

const meta: Meta<typeof LongValue> = {
  title: 'SIGTA/LongValue',
  component: LongValue,
  parameters: {
    docs: {
      description: {
        component: 'Un texto largo que escribió el titular, con sus renglones, plegado con "Ver todo" cuando no entra.',
      },
    },
  },
}

export default meta

export const Short: StoryObj<typeof LongValue> = {
  name: 'Corto',
  args: { text: 'Cursé la asignatura en otra facultad de la UBA.' },
}

export const Long: StoryObj<typeof LongValue> = {
  name: 'Largo',
  args: {
    text: Array.from(
      { length: 12 },
      (_, index) => `Renglón ${index + 1} del detalle, con el texto tal como lo escribió el titular.`,
    ).join('\n'),
  },
}
