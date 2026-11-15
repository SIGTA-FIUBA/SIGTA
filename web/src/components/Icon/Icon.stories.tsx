import { Group, Stack, Text } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { Icon, type IconName } from './Icon'

const meta: Meta = {
  title: 'SIGTA/Icon',
  parameters: {
    docs: {
      description: {
        component:
          'Los íconos de la interfaz: un solo set, de contorno, con el trazo de Tabler Icons y 1,5 px de grosor a cualquier tamaño (`layout.iconStroke` en `tokens.json`). Miden 20 en la navegación y 16 junto a un texto, y toman el color del texto que los rodea. Un ícono nuevo se agrega en `Icon.tsx`.',
      },
    },
  },
}

export default meta

const names: IconName[] = ['home', 'inbox', 'catalog', 'chevron-right', 'chevron-down']

export const AllIcons: StoryObj = {
  name: 'Icon',
  render: () => (
    <Stack gap="lg">
      {([20, 16] as const).map((size) => (
        <Group key={size} gap="xl">
          {names.map((name) => (
            <Stack key={name} gap="xs" align="center">
              <Icon name={name} size={size} />
              <Text size="xs" ff="monospace">
                {name} {size}
              </Text>
            </Stack>
          ))}
        </Group>
      ))}
    </Stack>
  ),
}

export const InheritsColor: StoryObj = {
  name: 'Icon, color del texto',
  render: () => (
    <Group gap="xl">
      <Text c="fiuba.7">
        <Icon name="inbox" />
      </Text>
      <Text c="dimmed">
        <Icon name="inbox" />
      </Text>
    </Group>
  ),
}
