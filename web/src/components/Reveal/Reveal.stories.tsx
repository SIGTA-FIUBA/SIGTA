import { Paper, Stack, Text } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { Reveal } from './Reveal'

const meta: Meta<typeof Reveal> = {
  title: 'SIGTA/Reveal',
  component: Reveal,
  parameters: {
    docs: {
      description: {
        component:
          'Hace entrar el contenido cuando llega la página: sube `motion.rise` y aparece en `motion.slow`, una sola vez. En una lista, cada elemento entra `motion.stagger` después del anterior, hasta `motion.staggerLimit`. Con movimiento reducido solo aparece. Para volver a verlo, recargá la historia.',
      },
    },
  },
}

export default meta

export const List: StoryObj<typeof Reveal> = {
  name: 'Reveal, en una lista',
  render: () => (
    <Stack gap="xs">
      {['Equivalencia entre asignaturas', 'Reconocimiento de créditos', 'Pase de carrera'].map((name, index) => (
        <Reveal key={name} index={index}>
          <Paper withBorder p="md">
            <Text fw={600}>{name}</Text>
          </Paper>
        </Reveal>
      ))}
    </Stack>
  ),
}
