import type { Meta, StoryObj } from '@storybook/react-vite'
import { Anchor, Text } from '@mantine/core'
import { PageSection } from './PageSection'

const meta: Meta = {
  title: 'SIGTA/PageSection',
  parameters: {
    docs: {
      description: {
        component:
          'Una sección de una página sobre la grilla de doce columnas: en pantalla ancha, el título, la bajada y una acción opcional en las tres primeras columnas y el contenido en las otras nueve, así todas las secciones empiezan en la misma línea; en el celular, el título arriba. La usan el inicio y la portada.',
      },
    },
  },
}

export default meta

export const Default: StoryObj = {
  name: 'PageSection',
  render: () => (
    <PageSection
      id="story-section"
      title="Actividad reciente"
      lead="Dónde quedó cada trámite después de su último paso."
      action={<Anchor href="#">Ver la bandeja</Anchor>}
    >
      <Text>Titular Uno inició el trámite.</Text>
    </PageSection>
  ),
}
