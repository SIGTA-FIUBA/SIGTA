import type { Meta, StoryObj } from '@storybook/react-vite'
import { GoogleSignInButton } from './GoogleSignInButton'

const meta: Meta = {
  title: 'SIGTA/GoogleSignInButton',
  parameters: {
    docs: {
      description: {
        component:
          'El ingreso con la cuenta de la UBA. Es la única excepción a las reglas del tema: Google exige sus colores y su tipografía, así que tiene su propio CSS.',
      },
    },
  },
}

export default meta

export const Default: StoryObj = {
  name: 'GoogleSignInButton',
  render: () => <GoogleSignInButton />,
}
