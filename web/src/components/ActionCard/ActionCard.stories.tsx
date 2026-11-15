import { SimpleGrid } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { MemoryRouter } from 'react-router'
import { ActionCard } from './ActionCard'

const meta: Meta<typeof ActionCard> = {
  title: 'SIGTA/ActionCard',
  component: ActionCard,
  decorators: [
    (Story) => (
      <MemoryRouter>
        <Story />
      </MemoryRouter>
    ),
  ],
  parameters: {
    docs: {
      description: {
        component:
          'La única caja del contenido, y existe solo porque toda la tarjeta es un enlace: borde `border-subtle`, radio 10, relleno 24. Con el puntero, fondo `surface-hover` y borde `border-control`. Nunca una sola en una fila: un acceso solo es un botón.',
      },
    },
  },
}

export default meta

export const Cards: StoryObj<typeof ActionCard> = {
  name: 'ActionCard',
  render: () => (
    <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="lg">
      <ActionCard to="/inbox" title="Bandeja de trámites" description="Trámites en curso y acciones pendientes." />
      <ActionCard to="/catalog" title="Catálogo de trámites" description="Requisitos, plazos y documentación." />
    </SimpleGrid>
  ),
}
