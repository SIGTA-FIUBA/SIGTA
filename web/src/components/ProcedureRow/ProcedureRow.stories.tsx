import type { Meta, StoryObj } from '@storybook/react-vite'
import { MemoryRouter } from 'react-router'
import { ProcedureRow, ProcedureRows } from './ProcedureRow'

const meta: Meta = {
  title: 'SIGTA/ProcedureRow',
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
          'Un trámite en una lista: el nombre, su estado, el próximo paso y la fecha de la última novedad. Toda la fila es un enlace; con el puntero, fondo `surface-hover`. Las filas viven en un solo contenedor con radio 10 y un divisor entre ellas.',
      },
    },
  },
}

export default meta

export const Rows: StoryObj = {
  name: 'ProcedureRow',
  render: () => (
    <ProcedureRows label="Trámites">
      <ProcedureRow
        to="/procedures/1"
        name="Equivalencia entre asignaturas"
        detail="Martín Sosa"
        status={{ tone: 'office', label: 'En la Secretaría' }}
        nextStep="Revisión de la Secretaría"
        date="Hoy"
      />
      <ProcedureRow
        to="/procedures/2"
        name="Reconocimiento de créditos"
        detail="Ana Ruiz"
        status={{ tone: 'external', label: 'En evaluación externa' }}
        nextStep="Dictamen del departamento"
        date="Hace 4 días"
      />
      <ProcedureRow
        to="/procedures/3"
        name="Equivalencia entre asignaturas"
        detail="Julia Paz"
        status={{ tone: 'applicant', label: 'Pendiente del titular' }}
        nextStep="Corrección de la solicitud"
        date="Hace 9 días"
      />
    </ProcedureRows>
  ),
}
