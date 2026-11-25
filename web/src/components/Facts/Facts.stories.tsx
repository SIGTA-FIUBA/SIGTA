import type { Meta, StoryObj } from '@storybook/react-vite'
import { Text } from '@mantine/core'
import { Facts } from './Facts'
import { Fact } from '../Fact/Fact'

const meta: Meta = {
  title: 'SIGTA/Facts',
  parameters: {
    docs: {
      description: {
        component:
          'Lo que algo es, como una lista de definiciones entre filetes y sin caja: un dato debajo del otro con un filete fino entre ellos. Partidos, el rótulo toma el primer tercio de la línea y el valor el resto; apilados, el rótulo va arriba, para una columna angosta.',
      },
    },
  },
}

export default meta

export const Split: StoryObj = {
  name: 'Partidos',
  render: () => (
    <Facts label="Datos del trámite">
      <Fact label="Titular" split>
        <Text>Ana Ruiz</Text>
      </Fact>
      <Fact label="Inicio" split>
        <Text>02/11/2026</Text>
      </Fact>
      <Fact label="A cargo de" split>
        <Text>Dirección de Alumnos</Text>
      </Fact>
    </Facts>
  ),
}

export const Stacked: StoryObj = {
  name: 'Apilados',
  render: () => (
    <Facts label="Resumen del trámite">
      <Fact label="Paso">
        <Text>Revisar la solicitud</Text>
      </Fact>
      <Fact label="A cargo de">
        <Text>Dirección de Alumnos</Text>
      </Fact>
    </Facts>
  ),
}
