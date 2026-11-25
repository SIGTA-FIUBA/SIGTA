import { Button, Input } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { tokens } from '../../theme/tokens'
import { Icon } from '../Icon/Icon'
import { SearchShortcut } from './SearchShortcut'

const meta: Meta = {
  title: 'SIGTA/SearchShortcut',
  parameters: {
    docs: {
      description: {
        component:
          'El atajo de la búsqueda, junto al botón que la abre: Cmd K en Mac y Ctrl K en el resto, en caption a 500 y en el color secundario de donde está. Dentro del campo que abre la búsqueda va como tecla (`keycap`), en el `Kbd` de Mantine.',
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

export const Keycap: StoryObj = {
  name: 'SearchShortcut keycap',
  render: () => (
    <Input
      component="button"
      pointer
      w={tokens.layout.search}
      leftSection={<Icon name="search" size={16} />}
      leftSectionPointerEvents="none"
      rightSection={<SearchShortcut keycap />}
      rightSectionWidth={tokens.layout.searchShortcut}
      rightSectionPointerEvents="none"
    >
      <Input.Placeholder>Buscar un trámite</Input.Placeholder>
    </Input>
  ),
}
