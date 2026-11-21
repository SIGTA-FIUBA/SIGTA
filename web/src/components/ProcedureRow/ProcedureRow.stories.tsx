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
          'Un trámite en una lista, en cuatro zonas: qué es (el tipo y, en caption `ink-500`, el expediente), de quién es (el titular y su carrera, en body-sm `ink-600`), en qué etapa está (StepBar y el paso abierto) y quién actúa (el estado y "A cargo de" el área, en caption `ink-600`). A la derecha, la última novedad y, si hay, el vencimiento en `warning-700`; al final, un chevron de 16 en `ink-500`. Un punto `brand-500` marca los trámites con novedades. Toda la fila es un enlace: con el puntero, fondo `surface-hover` y el chevron en `action-600`, 2 a la derecha; presionada, un tono más oscuro. Las filas viven en un solo contenedor blanco con radio 10 y un divisor entre ellas.',
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
        fileNumber="EX-2026-00012345"
        applicant="Martín Sosa"
        program="Ingeniería en Informática"
        step={{ number: 3, count: 9, name: 'Revisión de la Secretaría' }}
        area="Secretaría Académica"
        status={{ tone: 'office', label: 'En un área' }}
        date="Hoy"
        news
      />
      <ProcedureRow
        to="/procedures/2"
        type="Reconocimiento de créditos"
        applicant="Ana Ruiz"
        step={{ number: 6, count: 9, name: 'Dictamen del departamento' }}
        area="Departamento de Computación"
        status={{ tone: 'external', label: 'Derivado' }}
        date="Hace 4 días"
        due="el 20/11"
      />
      <ProcedureRow
        to="/procedures/3"
        type="Equivalencia entre asignaturas"
        fileNumber="EX-2026-00012388"
        applicant="Julia Paz"
        program="Ingeniería Civil"
        step={{ number: 2, count: 9, name: 'Corrección de la solicitud' }}
        area="Titular"
        status={{ tone: 'applicant', label: 'Pendiente del titular' }}
        date="Hace 9 días"
      />
    </ProcedureRows>
  ),
}
