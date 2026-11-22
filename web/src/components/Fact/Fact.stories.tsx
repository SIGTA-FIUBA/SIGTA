import type { Meta, StoryObj } from '@storybook/react-vite'
import { Text } from '@mantine/core'
import { Fact } from './Fact'
import { Facts } from '../Facts/Facts'

const meta: Meta = {
  title: 'SIGTA/Fact',
  parameters: {
    docs: {
      description: {
        component:
          'Un dato de Facts: su rótulo como dt, en body-sm `ink-600`, y su valor como dd. Partido, el rótulo toma cuatro de doce columnas desde sm; si no, va arriba del valor.',
      },
    },
  },
}

export default meta

export const Default: StoryObj = {
  name: 'Fact',
  render: () => (
    <Facts label="Datos">
      <Fact label="Plazos" split>
        <Text>La Secretaría responde en diez días hábiles.</Text>
      </Fact>
    </Facts>
  ),
}
