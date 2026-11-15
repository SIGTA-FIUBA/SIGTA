import { Group, Text } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { NewsDot } from './NewsDot'

const meta: Meta = {
  title: 'SIGTA/NewsDot',
  parameters: {
    docs: {
      description: {
        component:
          'Un punto de 8 en `brand-500` junto al título de un trámite que se movió desde la última vez que la persona lo vio. Para un lector de pantalla dice "Con novedades".',
      },
    },
  },
}

export default meta

export const Dot: StoryObj = {
  name: 'NewsDot',
  render: () => (
    <Group gap="xs" wrap="nowrap">
      <NewsDot />
      <Text fw={600}>Equivalencia entre asignaturas</Text>
    </Group>
  ),
}
