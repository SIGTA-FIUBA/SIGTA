import type { Meta, StoryObj } from '@storybook/react-vite'
import { Badge } from '@mantine/core'
import { Placed } from './Placed'

const meta: Meta = {
  title: 'SIGTA/Placed',
  parameters: {
    docs: {
      description: {
        component:
          'Pone lo que contiene en un punto del diagrama de pasos, en píxeles calculados por stepLayout. Es el único elemento con estilo en línea (ADR-015): una posición es una coordenada, no un valor de diseño.',
      },
    },
  },
}

export default meta

export const Points: StoryObj = {
  name: 'Puntos en un marco',
  render: () => (
    <Placed frame width={320} height={120}>
      <Placed x={0} y={0}>
        <Badge>Inicio</Badge>
      </Placed>
      <Placed x={160} y={48} anchor="middle">
        <Badge>Centro</Badge>
      </Placed>
      <Placed x={320} y={96} anchor="end">
        <Badge>Fin</Badge>
      </Placed>
    </Placed>
  ),
}
