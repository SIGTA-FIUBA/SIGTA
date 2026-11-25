import type { Meta, StoryObj } from '@storybook/react-vite'
import { BrandSignature } from './BrandSignature'

const meta: Meta = {
  title: 'SIGTA/BrandSignature',
  parameters: {
    docs: {
      description: {
        component:
          'La firma conjunta: el logo de SIGTA, un divisor y el logo de la FIUBA, siempre sobre blanco. Va en la pantalla de ingreso, en su pie y en "Acerca de SIGTA"; en el panel alcanza con el logo de SIGTA. El logo de la FIUBA es un PNG provisorio, a reemplazar por el SVG oficial. La firma tiene que pasar por Comunicación Institucional.',
      },
    },
  },
}

export default meta

export const Signature: StoryObj = {
  name: 'BrandSignature',
  render: () => <BrandSignature />,
}
