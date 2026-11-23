import type { Meta, StoryObj } from '@storybook/react-vite'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { PendingRow } from './PendingRow'
import { ProcedureRows } from '../ProcedureRows/ProcedureRows'

const meta: Meta = {
  title: 'SIGTA/PendingRow',
  decorators: [(Story) => <RouterProvider router={createMemoryRouter([{ path: '*', element: <Story /> }])} />],
  parameters: {
    docs: {
      description: {
        component:
          'Algo para hacer, en una lista de filas: arriba lo que hay que hacer (body semibold), abajo el trámite y, para el personal, su titular (body-sm `ink-600`), y a la derecha el plazo si lo hay. Toda la fila es un enlace, sin caja; la marca gris de la lista la sigue con el puntero.',
      },
    },
  },
}

export default meta

export const Rows: StoryObj = {
  name: 'PendingRow',
  render: () => (
    <ProcedureRows label="Pendientes">
      <PendingRow to="/procedures/1" task="Certificado analítico firmado" context="Equivalencia entre asignaturas" />
      <PendingRow
        to="/procedures/2"
        task="Revisión de la documentación"
        context="Reconocimiento de créditos, Laura Méndez"
        due="Vence en 3 días"
      />
    </ProcedureRows>
  ),
}
