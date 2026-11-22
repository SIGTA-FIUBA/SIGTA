import '@fontsource-variable/inter/wght.css'
import '@mantine/core/styles.css'
import { MantineProvider } from '@mantine/core'
import type { Preview } from '@storybook/react-vite'
import { MotionConfig } from 'motion/react'
import { cssVariablesResolver, theme } from '../src/theme/theme'
import '../src/theme/global.css'

const preview: Preview = {
  decorators: [
    (Story) => (
      <MantineProvider theme={theme} cssVariablesResolver={cssVariablesResolver} forceColorScheme="light">
        <MotionConfig reducedMotion="user">
          <Story />
        </MotionConfig>
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
