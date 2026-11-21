import {
  ActionIcon,
  Alert,
  Button,
  EmptyState,
  Group,
  Indicator,
  Loader,
  Notification,
  Progress,
  RingProgress,
  SemiCircleProgress,
  Skeleton,
  Stack,
  Text,
} from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { MenuIcon } from '../../layout/MenuIcon'

// Feedback: messages, loading states and progress. Each story shows the
// component as SIGTA uses it; the controls of the first one try every property.

const meta: Meta = {
  title: 'Componentes/Avisos',
  parameters: {
    docs: {
      description: {
        component:
          'Toda pantalla que muestra datos contempla cuatro estados: cargando (`Loader`, `Skeleton`), vacío (`EmptyState`), error (`Alert color="red"`) y con datos. El rojo se usa solo para errores.',
      },
    },
  },
}

export default meta

export const AlertStory: StoryObj<typeof Alert> = {
  name: 'Alert',
  render: (args) => <Alert {...args} />,
  args: {
    title: 'Documentación incompleta',
    children: 'Falta el programa de la asignatura de origen. Adjuntalo para continuar con la solicitud.',
    variant: 'light',
    withCloseButton: false,
  },
  argTypes: {
    variant: { control: 'select', options: ['light', 'filled', 'outline', 'default', 'transparent', 'white'] },
    color: { control: 'select', options: ['fiuba', 'gray', 'orange', 'red'] },
    radius: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
  },
  parameters: {
    docs: {
      description: {
        story:
          'Un aviso dentro de la pantalla. En el primario para informar; en rojo solo cuando algo falló, y dice qué hacer.',
      },
    },
  },
}

export const AlertError: StoryObj = {
  name: 'Alert, error',
  render: () => (
    <Stack gap="sm">
      <Alert role="alert">La bandeja de trámites está disponible solo para el personal de la Secretaría.</Alert>
      <Alert color="red" role="alert" title="No se pudo cargar la bandeja de trámites">
        Volvé a intentarlo más tarde.
      </Alert>
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Como en la bandeja: falta de permiso en el primario, error de carga en rojo. Llevan `role="alert"` para que el lector de pantalla los anuncie.',
      },
    },
  },
}

export const NotificationStory: StoryObj = {
  name: 'Notification',
  render: () => (
    <Stack gap="sm">
      <Notification title="Solicitud enviada" withCloseButton={false}>
        La equivalencia de Análisis Matemático II quedó en la Secretaría.
      </Notification>
      <Notification loading title="Enviando la solicitud" withCloseButton={false}>
        Se están adjuntando los documentos.
      </Notification>
      <Notification color="red" title="No se pudo enviar la solicitud" closeButtonProps={{ 'aria-label': 'Cerrar' }}>
        Revisá la conexión y volvé a intentarlo.
      </Notification>
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Confirma una acción que terminó o que está en curso, fuera del contenido de la pantalla.',
      },
    },
  },
}

export const LoaderStory: StoryObj = {
  name: 'Loader',
  render: () => (
    <Group align="center">
      <Loader aria-label="Cargando los trámites" />
      <Loader type="dots" aria-label="Cargando" />
      <Loader size="xs" aria-label="Cargando" />
    </Group>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'El estado de carga de una pantalla, como en la bandeja. Lleva `aria-label` con lo que se está cargando.',
      },
    },
  },
}

export const ProgressStory: StoryObj = {
  name: 'Progress',
  render: () => (
    <Stack gap="md">
      <Progress value={60} aria-label="Documentación adjunta" />
      <Progress.Root size="xl">
        <Progress.Section value={50} color="fiuba">
          <Progress.Label>En un área</Progress.Label>
        </Progress.Section>
        <Progress.Section value={30} color="orange">
          <Progress.Label>Derivado</Progress.Label>
        </Progress.Section>
        <Progress.Section value={20} color="gray">
          <Progress.Label>Pendiente del titular</Progress.Label>
        </Progress.Section>
      </Progress.Root>
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story: 'El avance de una tarea, o la proporción de trámites en cada estado con los colores de estado.',
      },
    },
  },
}

export const RingProgressStory: StoryObj = {
  name: 'RingProgress',
  render: () => (
    <RingProgress
      sections={[
        { value: 50, color: 'fiuba', tooltip: 'En un área' },
        { value: 30, color: 'orange', tooltip: 'Derivado' },
        { value: 20, color: 'gray', tooltip: 'Pendiente del titular' },
      ]}
      label={
        <Text ta="center" size="sm" fw={700}>
          24 trámites
        </Text>
      }
    />
  ),
  parameters: {
    docs: {
      description: { story: 'Una proporción en un resumen. Para un número exacto, mejor el texto.' },
    },
  },
}

export const SemiCircleProgressStory: StoryObj = {
  name: 'SemiCircleProgress',
  render: () => (
    <SemiCircleProgress value={75} label="Créditos reconocidos: 75 %" filledSegmentColor="fiuba" size={240} />
  ),
  parameters: {
    docs: { description: { story: 'Un avance hacia una meta, como los créditos reconocidos de una carrera.' } },
  },
}

export const SkeletonStory: StoryObj = {
  name: 'Skeleton',
  render: () => (
    <Stack gap="sm">
      <Skeleton height={12} width="40%" />
      <Skeleton height={8} />
      <Skeleton height={8} />
      <Skeleton height={8} width="70%" />
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Reserva el lugar del contenido mientras carga, cuando se conoce su forma (por ejemplo, filas de la tabla).',
      },
    },
  },
}

export const EmptyStateStory: StoryObj = {
  name: 'EmptyState',
  render: () => (
    <EmptyState
      icon={<MenuIcon name="inbox" />}
      variant="light"
      title="No hay trámites que requieran acción de este perfil."
      description="Los trámites asignados aparecen acá cuando llegan a una instancia de este perfil."
    >
      <EmptyState.Actions>
        <Button variant="default">Ver todos los trámites</Button>
      </EmptyState.Actions>
    </EmptyState>
  ),
  parameters: {
    docs: {
      description: {
        story: 'El estado vacío de una lista: dice por qué no hay nada y, si se puede, qué hacer.',
      },
    },
  },
}

export const IndicatorStory: StoryObj = {
  name: 'Indicator',
  render: () => (
    <Group gap="xl">
      <Indicator label={3} size={16}>
        <ActionIcon variant="default" size="lg" aria-label="Bandeja de trámites, 3 con acción requerida">
          <MenuIcon name="inbox" />
        </ActionIcon>
      </Indicator>
      <Indicator processing>
        <ActionIcon variant="default" size="lg" aria-label="Catálogo de trámites, con novedades">
          <MenuIcon name="catalog" />
        </ActionIcon>
      </Indicator>
    </Group>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Una marca sobre otro elemento, como la cantidad de trámites con acción requerida. El número también va en el `aria-label`.',
      },
    },
  },
}
