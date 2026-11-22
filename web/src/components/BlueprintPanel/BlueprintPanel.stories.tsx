import { Stack, Text, Title } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { BlueprintPanel } from './BlueprintPanel'

const meta: Meta<typeof BlueprintPanel> = {
  title: 'SIGTA/BlueprintPanel',
  component: BlueprintPanel,
  parameters: {
    docs: {
      description: {
        component:
          'El panel azul del logo con la trama de un plano: la cabecera del Inicio y de la página pública. Con el mouse encima, la trama se enciende alrededor del puntero; con el dedo o con movimiento reducido queda quieta. El texto va en blanco o en `fiuba.1` y conserva su contraste sin la trama.',
      },
    },
  },
}

export default meta

export const Header: StoryObj<typeof BlueprintPanel> = {
  name: 'BlueprintPanel, cabecera',
  render: () => (
    <BlueprintPanel component="header" p="xl" radius="md">
      <Stack gap="xs">
        <Text size="sm" c="fiuba.1">
          Sábado, 21 de noviembre
        </Text>
        <Title order={1} c="white">
          Hola, Ana
        </Title>
      </Stack>
    </BlueprintPanel>
  ),
}
