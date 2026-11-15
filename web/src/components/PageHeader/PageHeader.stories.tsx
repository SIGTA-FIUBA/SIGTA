import { Button } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { PageHeader } from './PageHeader'

const meta: Meta<typeof PageHeader> = {
  title: 'SIGTA/PageHeader',
  component: PageHeader,
  parameters: {
    docs: {
      description: {
        component:
          'Abre cada página: el título en `title-page`, una descripción opcional de hasta 68 caracteres por línea y, a la derecha, la acción principal, el único botón primario de la pantalla. Sin rótulo encima: el logo ya dice dónde se está.',
      },
    },
  },
}

export default meta

export const WithAction: StoryObj<typeof PageHeader> = {
  name: 'PageHeader, con acción',
  args: { title: 'Inicio', action: <Button>Iniciar un trámite</Button> },
}

export const WithDescription: StoryObj<typeof PageHeader> = {
  name: 'PageHeader, con descripción',
  args: {
    title: 'Catálogo de trámites',
    description: 'Requisitos, plazos y documentación de cada trámite que se realiza en la Secretaría.',
  },
}
