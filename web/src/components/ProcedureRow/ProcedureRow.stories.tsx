import type { Meta, StoryObj } from '@storybook/react-vite'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { ProcedureRow } from './ProcedureRow'
import { ProcedureRows } from '../ProcedureRows/ProcedureRows'

const meta: Meta = {
  title: 'SIGTA/ProcedureRow',
  decorators: [
    // The rows follow the opening of a page (useViewTransitionState), which
    // takes a data router.
    (Story) => <RouterProvider router={createMemoryRouter([{ path: '*', element: <Story /> }])} />,
  ],
  parameters: {
    docs: {
      description: {
        component:
          'Un trámite en una lista, sin caja, en cuatro columnas: qué es y de quién (el tipo en body semibold, cortado con el nombre completo al pasar el puntero, y el titular en body-sm `ink-600`), en qué paso está, quién actúa ("A cargo de" el área, o el estado en las listas propias) y la última novedad, todo en body-sm. En el celular las columnas se apilan y la tercera dice su rótulo; en pantalla ancha lo dice el encabezado de la columna. Toda la fila es un enlace: con el puntero o el teclado, la marca gris de la lista se desliza hasta ella y aparece un chevron de 16 en `action-600`; presionada, la marca se oscurece. Al abrirla, el tipo pasa a ser el título de la página.',
      },
    },
  },
}

export default meta

export const Rows: StoryObj = {
  name: 'ProcedureRow',
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
        step={{ name: 'Dictamen del departamento' }}
        standing={{ label: 'A cargo de', text: 'Departamento de Computación' }}
        date="Hace 4 días"
      />
      <ProcedureRow
        to="/procedures/3"
        type="Equivalencia entre asignaturas"
        applicant="Julia Paz"
        step={{ name: 'Corrección de la solicitud' }}
        standing={{ text: 'Pendiente del titular' }}
        date="Hace 9 días"
      />
    </ProcedureRows>
  ),
}
