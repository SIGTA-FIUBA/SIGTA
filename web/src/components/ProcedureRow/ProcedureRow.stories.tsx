import type { Meta, StoryObj } from '@storybook/react-vite'
import { MemoryRouter } from 'react-router'
import { ProcedureRow } from './ProcedureRow'
import { ProcedureRows } from '../ProcedureRows/ProcedureRows'

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
          'Un trámite en una lista, en cuatro zonas: qué es (el tipo, cortado con el nombre completo al pasar el puntero), de quién es (el titular, en body-sm `ink-600`), en qué etapa está (el paso abierto) y quién actúa (el estado y "A cargo de" el área, en caption `ink-600`). A la derecha, la última novedad; al final, un chevron de 16 en `ink-500`. Toda la fila es un enlace: con el puntero, fondo `surface-hover` y el chevron en `action-600`, 2 a la derecha; presionada, un tono más oscuro. Las filas viven en un solo contenedor blanco (`ProcedureRows`) con radio 10 y un divisor entre ellas.',
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
        step={{ name: 'Dictamen del departamento' }}
        area="Departamento de Computación"
        status={{ tone: 'external', label: 'Derivado' }}
        date="Hace 4 días"
      />
      <ProcedureRow
        to="/procedures/3"
        type="Equivalencia entre asignaturas"
        applicant="Julia Paz"
        step={{ name: 'Corrección de la solicitud' }}
        status={{ tone: 'applicant', label: 'Pendiente del titular' }}
        date="Hace 9 días"
      />
    </ProcedureRows>
  ),
}
