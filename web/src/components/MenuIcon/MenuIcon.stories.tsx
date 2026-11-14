import { Group, Stack, Text } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { MenuIcon, type MenuIconName } from './MenuIcon'

const meta: Meta = {
  title: 'SIGTA/MenuIcon',
  parameters: {
    docs: {
      description: {
        component:
          'Los íconos del menú lateral, con el trazo de Tabler Icons. Toman el color del texto que los rodea. Un ícono nuevo se agrega en `MenuIcon.tsx`.',
      },
    },
  },
}

export default meta

const names: MenuIconName[] = ['home', 'inbox', 'catalog']

export const AllIcons: StoryObj = {
  name: 'MenuIcon',
  render: () => (
    <Group gap="xl">
      {names.map((name) => (
        <Stack key={name} gap="xs" align="center">
          <MenuIcon name={name} />
          <Text size="xs" ff="monospace">
            {name}
          </Text>
        </Stack>
      ))}
    </Group>
  ),
}

export const InheritsColor: StoryObj = {
  name: 'MenuIcon, color del texto',
  render: () => (
    <Group gap="xl">
      <Text c="fiuba">
        <MenuIcon name="inbox" />
      </Text>
      <Text c="dimmed">
        <MenuIcon name="inbox" />
      </Text>
    </Group>
  ),
}
