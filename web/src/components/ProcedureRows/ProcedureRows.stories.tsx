import type { Meta, StoryObj } from '@storybook/react-vite'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { ProcedureRow } from '../ProcedureRow/ProcedureRow'
import { ProcedureRows } from './ProcedureRows'

const meta: Meta = {
  title: 'SIGTA/ProcedureRows',
  decorators: [
    // The rows follow the opening of a page (useViewTransitionState), which
    // takes a data router.
    (Story) => <RouterProvider router={createMemoryRouter([{ path: '*', element: <Story /> }])} />,
  ],
  parameters: {
    docs: {
      description: {
        component:
          'Una lista de trámites como una tabla de libro, sin caja: un filete arriba, uno entre filas y uno abajo, en `gray-3`, y en pantalla ancha el encabezado de las cuatro columnas en caption `ink-600`. Una sola marca gris sigue al puntero o al foco de fila en fila. Cada hijo es una fila (`ProcedureRow`).',
      },
    },
  },
}

export default meta

export const Rows: StoryObj = {
  name: 'ProcedureRows',
  render: () => (
    <ProcedureRows label="Trámites" columns={['Trámite', 'Paso', 'A cargo de', 'Último movimiento']}>
      <ProcedureRow
        to="/procedures/1"
        type="Equivalencia entre asignaturas"
        applicant="Martín Sosa"
        step={{ name: 'Revisión de la Secretaría' }}
        standing={{ label: 'A cargo de', text: 'Secretaría de Gestión Académica' }}
        date="Hoy"
      />
      <ProcedureRow
        to="/procedures/2"
        type="Reconocimiento de créditos"
        applicant="Ana Ruiz"
        step={{ name: 'Corrección de la solicitud' }}
        standing={{ text: 'Pendiente del titular' }}
        date="Hace 9 días"
      />
    </ProcedureRows>
  ),
}
