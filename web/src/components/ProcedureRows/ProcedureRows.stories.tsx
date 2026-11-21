import type { Meta, StoryObj } from '@storybook/react-vite'
import { MemoryRouter } from 'react-router'
import { ProcedureRow } from '../ProcedureRow/ProcedureRow'
import { ProcedureRows } from './ProcedureRows'

const meta: Meta = {
  title: 'SIGTA/ProcedureRows',
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
          'El contenedor de una lista de trámites: una lista con nombre, blanca, con borde, radio 10 y un divisor entre filas. Cada hijo es una fila (`ProcedureRow`).',
      },
    },
  },
}

export default meta

export const Rows: StoryObj = {
  name: 'ProcedureRows',
  render: () => (
    <ProcedureRows label="Trámites">
      <ProcedureRow
        to="/procedures/1"
        type="Equivalencia entre asignaturas"
        applicant="Martín Sosa"
        step={{ name: 'Revisión de la Secretaría' }}
        area="Secretaría de Gestión Académica"
        status={{ tone: 'office', label: 'En un área' }}
        date="Hoy"
      />
      <ProcedureRow
        to="/procedures/2"
        type="Reconocimiento de créditos"
        applicant="Ana Ruiz"
        step={{ name: 'Corrección de la solicitud' }}
        status={{ tone: 'applicant', label: 'Pendiente del titular' }}
        date="Hace 9 días"
      />
    </ProcedureRows>
  ),
}
