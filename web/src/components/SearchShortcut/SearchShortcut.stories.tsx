import { Button } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { Icon } from '../Icon/Icon'
import { SearchShortcut } from './SearchShortcut'

const meta: Meta = {
  title: 'SIGTA/SearchShortcut',
  parameters: {
    docs: {
      description: {
        component:
          'El atajo de la búsqueda, junto al botón que la abre: Cmd K en Mac y Ctrl K en el resto, en caption a 500 y en el color secundario de donde está.',
      },
    },
  },
}

export default meta

export const Shortcut: StoryObj = {
  name: 'SearchShortcut',
  render: () => (
    <Button variant="default" leftSection={<Icon name="search" size={16} />} rightSection={<SearchShortcut />}>
      Buscar en el catálogo
    </Button>
  ),
}
