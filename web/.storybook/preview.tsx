import '@mantine/core/styles.css'
import { MantineProvider } from '@mantine/core'
import type { Preview } from '@storybook/react-vite'
import { cssVariablesResolver, theme } from '../src/theme/theme'

const preview: Preview = {
  decorators: [
    (Story) => (
      <MantineProvider theme={theme} cssVariablesResolver={cssVariablesResolver} forceColorScheme="light">
        <Story />
      </MantineProvider>
    ),
  ],
  parameters: {
    layout: 'padded',
    options: {
      storySort: { order: ['Introducción', 'Marca', 'Fundamentos', 'Componentes', 'SIGTA'] },
    },
  },
  tags: ['autodocs'],
}

export default preview
