import { SimpleGrid } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { StartCard } from './StartCard'

const meta: Meta = {
  title: 'SIGTA/StartCard',
  decorators: [(Story) => <RouterProvider router={createMemoryRouter([{ path: '*', element: <Story /> }])} />],
  parameters: {
    docs: {
      description: {
        component:
          'Un tipo de trámite para iniciar, en el inicio: el nombre en body semibold, y la descripción en dos líneas a lo sumo; lo que pide queda para el formulario. Toda la tarjeta es un enlace (`Paper variant="interactive"`): con el puntero el fondo pasa a `surface-hover` y la flecha avanza en `action-600`. Van en una grilla de tres columnas que cierra las filas.',
      },
    },
  },
}

export default meta

export const Cards: StoryObj = {
  name: 'StartCard',
  render: () => (
    <SimpleGrid cols={{ base: 1, sm: 3 }}>
      <StartCard
        to="/catalog/1/start"
        name="Equivalencia entre asignaturas"
        description="Reconocimiento de una asignatura aprobada en otra carrera o universidad."
      />
      <StartCard
        to="/catalog/2/start"
        name="Certificado de alumno regular"
        description="Constancia de regularidad en la carrera."
      />
      <StartCard to="/catalog/3/start" name="Pase de carrera" />
    </SimpleGrid>
  ),
}
