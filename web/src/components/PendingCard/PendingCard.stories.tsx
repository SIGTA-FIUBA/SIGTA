import { SimpleGrid } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { PendingCard } from './PendingCard'

const meta: Meta = {
  title: 'SIGTA/PendingCard',
  decorators: [(Story) => <RouterProvider router={createMemoryRouter([{ path: '*', element: <Story /> }])} />],
  parameters: {
    docs: {
      description: {
        component:
          'Algo que el titular tiene que hacer, en el inicio, para que no pase de largo: un panel con la marca en el color de acción (`ThemeIcon` light), lo que hay que hacer (body semibold), el trámite (body-sm `ink-600`), y el plazo en el color de acción si lo hay. Toda la tarjeta es el enlace, sin botón (`Paper variant="interactive"`). Van en una grilla de dos columnas desde breakpoint.md.',
      },
    },
  },
}

export default meta

export const Cards: StoryObj = {
  name: 'PendingCard',
  render: () => (
    <SimpleGrid cols={{ base: 1, md: 2 }}>
      <PendingCard to="/procedures/1" task="Certificado analítico firmado" context="Equivalencia entre asignaturas" />
      <PendingCard
        to="/procedures/2"
        task="Programa de la asignatura"
        context="Reconocimiento de créditos"
        due="Vence en 3 días"
      />
    </SimpleGrid>
  ),
}
