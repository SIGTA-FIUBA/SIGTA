import { Button } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { EmptyState } from './EmptyState'

const meta: Meta<typeof EmptyState> = {
  title: 'SIGTA/EmptyState',
  component: EmptyState,
  parameters: {
    docs: {
      description: {
        component:
          'Ocupa el lugar de una lista vacía: un título, una línea y qué hacer a continuación. Sin ilustración ni caja. El botón va solo si la pantalla no tiene otro primario.',
      },
    },
  },
}

export default meta

export const WithAction: StoryObj<typeof EmptyState> = {
  name: 'EmptyState, con acción',
  args: {
    title: 'Todavía no iniciaste trámites',
    text: 'Elegí uno del catálogo para ver sus requisitos e iniciarlo.',
    action: <Button>Iniciar un trámite</Button>,
  },
}

export const WithoutAction: StoryObj<typeof EmptyState> = {
  name: 'EmptyState, sin acción',
  args: {
    title: 'No hay trámites pendientes',
    text: 'Cuando un trámite requiera acción del perfil, aparece acá.',
  },
}
