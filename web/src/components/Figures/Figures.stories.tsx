import type { Meta, StoryObj } from '@storybook/react-vite'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { Figures } from './Figures'
import { Figure } from '../Figure/Figure'

const meta: Meta = {
  title: 'SIGTA/Figures',
  decorators: [(Story) => <RouterProvider router={createMemoryRouter([{ path: '*', element: <Story /> }])} />],
  parameters: {
    docs: {
      description: {
        component:
          'Una fila de cifras tan ancha como su columna, entre dos filetes y con un filete fino entre las celdas, sin cajas. Sus columnas dividen a las cifras, así la fila siempre cierra: dos cifras en dos columnas, cuatro en dos en el celular y en cuatro desde md.',
      },
    },
  },
}

export default meta

export const Two: StoryObj = {
  name: 'Dos cifras',
  render: () => (
    <Figures label="Resumen">
      <Figure value={4} label="Requieren acción" to="/inbox" />
      <Figure value={11} label="En curso" to="/inbox?view=all" />
    </Figures>
  ),
}

export const Four: StoryObj = {
  name: 'Cuatro cifras',
  render: () => (
    <Figures label="Resumen por estado">
      <Figure value={4} label="Acción requerida" to="/inbox" />
      <Figure value={6} label="En un área" to="/inbox?state=in_office" />
      <Figure value={2} label="Derivados" to="/inbox?state=referred" />
      <Figure value={3} label="Pendientes del titular" to="/inbox?state=with_applicant" />
    </Figures>
  ),
}
