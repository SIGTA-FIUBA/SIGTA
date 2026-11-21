import {
  ActionIcon,
  Burger,
  Button,
  Chip,
  CloseButton,
  CopyButton,
  FileButton,
  Group,
  Stack,
  Text,
  UnstyledButton,
} from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { MenuIcon } from '../../layout/MenuIcon'

// Buttons and the controls that act on a click. Each story shows the component
// as SIGTA uses it; the controls of the first one try every property.

const meta: Meta = {
  title: 'Componentes/Botones',
  parameters: {
    docs: {
      description: {
        component:
          'Una pantalla tiene un solo botón principal (`variant="filled"`); el resto usa `default`, `light` o `subtle`. El radio sale de `component.button.radius` en `tokens.json`.',
      },
    },
  },
}

export default meta

export const ButtonStory: StoryObj<typeof Button> = {
  name: 'Button',
  render: (args) => <Button {...args} />,
  args: { children: 'Iniciar trámite', variant: 'filled', size: 'sm', disabled: false, loading: false },
  argTypes: {
    variant: {
      control: 'select',
      options: ['filled', 'light', 'outline', 'default', 'subtle', 'transparent', 'white'],
    },
    size: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
    color: { control: 'select', options: ['fiuba', 'gray', 'red'] },
    radius: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
  },
  parameters: {
    docs: {
      description: { story: 'El botón. Probá las propiedades en los controles; el tema fija el radio y el color.' },
    },
  },
}

export const ButtonVariants: StoryObj = {
  name: 'Button, variantes',
  render: () => (
    <Group>
      <Button>Enviar solicitud</Button>
      <Button variant="default">Guardar borrador</Button>
      <Button variant="light">Ver detalle</Button>
      <Button variant="subtle">Cancelar</Button>
      <Button variant="outline">Descargar</Button>
      <Button disabled>No disponible</Button>
      <Button loading>Enviando</Button>
    </Group>
  ),
  parameters: {
    docs: {
      description: {
        story:
          '`filled` es la acción principal de la pantalla, una sola. `default` y `light` para las secundarias, `subtle` para las de menor peso.',
      },
    },
  },
}

export const ButtonSizes: StoryObj = {
  name: 'Button, tamaños',
  render: () => (
    <Group align="center">
      {(['xs', 'sm', 'md', 'lg', 'xl'] as const).map((size) => (
        <Button key={size} size={size}>
          Tamaño {size}
        </Button>
      ))}
    </Group>
  ),
}

export const ButtonWithSections: StoryObj = {
  name: 'Button, con ícono',
  render: () => (
    <Group>
      <Button leftSection={<MenuIcon name="inbox" />}>Bandeja de trámites</Button>
      <Button variant="default" rightSection={<MenuIcon name="catalog" />}>
        Ver catálogo
      </Button>
      <Button fullWidth variant="light">
        Ancho completo
      </Button>
    </Group>
  ),
}

export const ButtonGroupStory: StoryObj = {
  name: 'Button.Group',
  render: () => (
    <Button.Group>
      <Button variant="default">Anterior</Button>
      <Button variant="default">Siguiente</Button>
    </Button.Group>
  ),
  parameters: { docs: { description: { story: 'Botones unidos que hacen acciones hermanas.' } } },
}

export const ActionIconStory: StoryObj = {
  name: 'ActionIcon',
  render: () => (
    <Group>
      {(['filled', 'light', 'default', 'subtle'] as const).map((variant) => (
        <ActionIcon key={variant} variant={variant} size="lg" aria-label="Ir al inicio">
          <MenuIcon name="home" />
        </ActionIcon>
      ))}
      <ActionIcon.Group>
        <ActionIcon variant="default" size="lg" aria-label="Inicio">
          <MenuIcon name="home" />
        </ActionIcon>
        <ActionIcon variant="default" size="lg" aria-label="Bandeja">
          <MenuIcon name="inbox" />
        </ActionIcon>
      </ActionIcon.Group>
    </Group>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Botón con solo un ícono. Siempre lleva `aria-label`: sin texto, el lector de pantalla no tiene qué leer.',
      },
    },
  },
}

export const CloseButtonStory: StoryObj = {
  name: 'CloseButton',
  render: () => <CloseButton aria-label="Cerrar" />,
}

export const CopyButtonStory: StoryObj = {
  name: 'CopyButton',
  render: () => (
    <CopyButton value="EQ-2026-00042">
      {({ copied, copy }) => (
        <Button variant="light" onClick={copy}>
          {copied ? 'Número copiado' : 'Copiar el número de trámite'}
        </Button>
      )}
    </CopyButton>
  ),
}

export const FileButtonStory: StoryObj = {
  name: 'FileButton',
  render: () => (
    <FileButton onChange={() => {}} accept="application/pdf">
      {(props) => (
        <Button variant="default" {...props}>
          Adjuntar programa de la asignatura
        </Button>
      )}
    </FileButton>
  ),
  parameters: {
    docs: {
      description: { story: 'Un botón que abre el selector de archivos. Para un campo de formulario, `FileInput`.' },
    },
  },
}

export const BurgerStory: StoryObj = {
  name: 'Burger',
  render: () => (
    <Group>
      <Burger opened={false} aria-label="Abrir el menú" />
      <Burger opened aria-label="Cerrar el menú" />
    </Group>
  ),
  parameters: { docs: { description: { story: 'Abre y cierra el menú lateral en el celular (`AppLayout`).' } } },
}

export const ChipStory: StoryObj = {
  name: 'Chip',
  render: () => (
    <Chip.Group multiple defaultValue={['in_office']}>
      <Group gap="xs">
        <Chip value="in_office">En la Secretaría</Chip>
        <Chip value="referred">En evaluación externa</Chip>
        <Chip value="with_applicant">Pendiente del titular</Chip>
      </Group>
    </Chip.Group>
  ),
  parameters: { docs: { description: { story: 'Filtros que se prenden y apagan.' } } },
}

export const UnstyledButtonStory: StoryObj = {
  name: 'UnstyledButton',
  render: () => (
    <Stack gap="xs">
      <UnstyledButton>
        <Text size="sm">Un área que se puede tocar, sin aspecto de botón.</Text>
      </UnstyledButton>
      <Text size="xs" c="dimmed">
        Para componentes propios. Para un panel que es un control, `Paper variant=&quot;interactive&quot;`.
      </Text>
    </Stack>
  ),
}
