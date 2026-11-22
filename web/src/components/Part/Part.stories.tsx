import type { Meta, StoryObj } from '@storybook/react-vite'
import { Text } from '@mantine/core'
import { Part } from './Part'

const meta: Meta = {
  title: 'SIGTA/Part',
  parameters: {
    docs: {
      description: {
        component:
          'Una parte de una página: un filete arriba, su título en h2, una bajada opcional en body-sm `ink-600` y lo que contiene, sin caja alrededor. Las partes se separan por aire, no por bordes.',
      },
    },
  },
}

export default meta

export const Default: StoryObj = {
  name: 'Part',
  render: () => (
    <Part title="Historial" lead="Cada acción del trámite, de la más reciente a la primera.">
      <Text>Titular Uno inició el trámite.</Text>
    </Part>
  ),
}
