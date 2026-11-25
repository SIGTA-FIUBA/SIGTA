import type { Meta, StoryObj } from '@storybook/react-vite'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { MyProcedureRows } from './MyProcedureRows'

const meta: Meta = {
  title: 'SIGTA/MyProcedureRows',
  decorators: [(Story) => <RouterProvider router={createMemoryRouter([{ path: '*', element: <Story /> }])} />],
  parameters: {
    docs: {
      description: {
        component:
          'Los trámites propios, en filas de `ProcedureRow`: sin titular, con el estado en la tercera columna y, si está cerrado, la fecha de cierre. Una subsanación pendiente se dice debajo del tipo.',
      },
    },
  },
}

export default meta

export const Rows: StoryObj = {
  name: 'MyProcedureRows',
  render: () => (
    <MyProcedureRows
      label="Mis trámites"
      from="/my-procedures"
      procedures={[
        {
          id: '1',
          procedure_type: 1,
          procedure_type_name: 'Equivalencia entre asignaturas',
          step_name: 'Presentación del titular',
          state: 'with_applicant',
          correction_pending: true,
          started_at: '2026-11-02T13:00:00Z',
          last_moved_at: '2026-11-18T13:00:00Z',
          resolutions: [],
        },
        {
          id: '2',
          procedure_type: 2,
          procedure_type_name: 'Certificado de alumno regular',
          state: 'closed',
          correction_pending: false,
          started_at: '2026-10-02T13:00:00Z',
          last_moved_at: '2026-10-05T13:00:00Z',
          resolutions: [],
        },
      ]}
    />
  ),
}
