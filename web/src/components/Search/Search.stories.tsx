import { Button } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { openSearch } from './openSearch'
import { Search } from './Search'

const meta: Meta = {
  title: 'SIGTA/Search',
  decorators: [(Story) => <RouterProvider router={createMemoryRouter([{ path: '*', element: <Story /> }])} />],
  parameters: {
    docs: {
      description: {
        component:
          'La búsqueda de toda la plataforma, como un spotlight: Ctrl K (Cmd K en Mac) o "/" la abren desde cualquier página, las flechas eligen, Enter abre y Esc cierra. Busca sin acentos ni mayúsculas en lo que la persona ya tiene (pendientes, sus trámites, el catálogo y las páginas), agrupado, y resalta lo que coincide. Antes de escribir muestra lo primero de cada grupo.',
      },
    },
  },
}

export default meta

export const Spotlight: StoryObj = {
  name: 'Search',
  render: () => (
    <>
      <Button variant="default" onClick={openSearch}>
        Buscar
      </Button>
      <Search
        groups={[
          {
            name: 'Pendientes',
            icon: 'file-pencil',
            items: [
              {
                id: 'p1',
                label: 'Certificado analítico firmado',
                description: 'Equivalencia entre asignaturas',
                to: '/procedures/1',
              },
            ],
          },
          {
            name: 'Iniciar un trámite',
            icon: 'plus',
            idle: 2,
            items: [
              { id: 'c1', label: 'Equivalencia entre asignaturas', to: '/catalog/1/start' },
              { id: 'c2', label: 'Certificado de alumno regular', to: '/catalog/2/start' },
              { id: 'c3', label: 'Pase de carrera', to: '/catalog/3/start' },
            ],
          },
          { name: 'Ir a', icon: 'arrow-right', items: [{ id: 'g1', label: 'Mis trámites', to: '/my-procedures' }] },
        ]}
      />
    </>
  ),
}
