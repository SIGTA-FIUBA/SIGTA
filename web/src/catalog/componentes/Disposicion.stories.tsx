import {
  AppShell,
  AspectRatio,
  Badge,
  Box,
  Button,
  Center,
  Collapse,
  Container,
  Divider,
  Flex,
  Grid,
  Group,
  NavLink,
  Paper,
  ScrollArea,
  Scroller,
  SimpleGrid,
  Space,
  Splitter,
  Stack,
  Text,
} from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'
import type { ReactNode } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { Icon } from '../../components/Icon/Icon'

// Components that place other components on the screen. Each item is a panel
// with a border, so the space between them shows the layout.

const meta: Meta = {
  title: 'Componentes/Disposición',
  parameters: {
    docs: {
      description: {
        component:
          'Las pantallas se arman con estos componentes, sin CSS propio. Los espacios salen de `spacing` en `tokens.json` (`xs` a `xl`), nunca de números sueltos.',
      },
    },
  },
}

export default meta

function Item({ children }: { children: ReactNode }) {
  return (
    <Paper withBorder p="md">
      <Text size="sm">{children}</Text>
    </Paper>
  )
}

const items = (
  <>
    <Item>Datos del titular</Item>
    <Item>Asignatura aprobada</Item>
    <Item>Asignatura solicitada</Item>
  </>
)

export const StackStory: StoryObj<typeof Stack> = {
  name: 'Stack',
  render: (args) => <Stack {...args} />,
  args: { gap: 'md', align: 'stretch', children: items },
  argTypes: {
    gap: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
    align: { control: 'select', options: ['stretch', 'flex-start', 'center', 'flex-end'] },
    children: { control: false },
  },
  parameters: {
    docs: {
      description: { story: 'Apila en columna las secciones de una página y los campos de un formulario.' },
    },
  },
}

export const GroupStory: StoryObj<typeof Group> = {
  name: 'Group',
  render: (args) => <Group {...args} />,
  args: { gap: 'md', justify: 'flex-start', grow: false, children: items },
  argTypes: {
    gap: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
    justify: { control: 'select', options: ['flex-start', 'center', 'space-between', 'flex-end'] },
    children: { control: false },
  },
  parameters: {
    docs: {
      description: {
        story: 'Pone en fila los botones de una acción y los datos cortos, como un estado junto a su fecha.',
      },
    },
  },
}

export const SimpleGridStory: StoryObj = {
  name: 'SimpleGrid',
  render: () => (
    <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="md">
      <Item>Con acción requerida</Item>
      <Item>En evaluación externa</Item>
      <Item>Pendientes del titular</Item>
    </SimpleGrid>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Columnas del mismo ancho, como los resúmenes de la bandeja. En el celular pasan a una sola columna.',
      },
    },
  },
}

export const GridStory: StoryObj = {
  name: 'Grid',
  render: () => (
    <Grid gap="md">
      <Grid.Col span={{ base: 12, sm: 8 }}>
        <Item>Detalle del trámite</Item>
      </Grid.Col>
      <Grid.Col span={{ base: 12, sm: 4 }}>
        <Item>Historial de movimientos</Item>
      </Grid.Col>
    </Grid>
  ),
  parameters: {
    docs: {
      description: { story: 'Columnas de distinto ancho, como el detalle de un trámite junto a su historial.' },
    },
  },
}

export const FlexStory: StoryObj<typeof Flex> = {
  name: 'Flex',
  render: (args) => <Flex {...args} />,
  args: { gap: 'md', direction: { base: 'column', sm: 'row' }, align: 'stretch', children: items },
  argTypes: {
    gap: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
    align: { control: 'select', options: ['stretch', 'flex-start', 'center', 'flex-end'] },
    children: { control: false },
  },
  parameters: {
    docs: {
      description: {
        story:
          'Cuando la dirección cambia con el ancho de pantalla. Para una fila o una columna fija, `Group` o `Stack`.',
      },
    },
  },
}

export const ContainerStory: StoryObj = {
  name: 'Container',
  render: () => (
    <Container size="xs" bg="fiuba.0" p="md">
      <Text size="sm">Un formulario de inicio de trámite, con un ancho que se lee cómodo.</Text>
    </Container>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Limita el ancho del contenido y lo centra, como en los formularios de inicio de trámite.',
      },
    },
  },
}

export const CenterStory: StoryObj = {
  name: 'Center',
  render: () => (
    <AspectRatio ratio={4}>
      <Center bg="fiuba.0">
        <Text size="sm">No hay trámites con acción requerida.</Text>
      </Center>
    </AspectRatio>
  ),
  parameters: {
    docs: {
      description: { story: 'Centra un contenido en su espacio, como el mensaje de una lista vacía.' },
    },
  },
}

export const SpaceStory: StoryObj = {
  name: 'Space',
  render: () => (
    <Box>
      <Item>Datos del titular</Item>
      <Space h="xl" />
      <Item>Documentación presentada</Item>
    </Box>
  ),
  parameters: {
    docs: {
      description: { story: 'Un espacio suelto entre dos bloques. Casi siempre alcanza con el `gap` de `Stack`.' },
    },
  },
}

export const DividerStory: StoryObj = {
  name: 'Divider',
  render: () => (
    <Stack gap="md">
      <Text size="sm">Datos del trámite</Text>
      <Divider />
      <Divider label="Documentación presentada" labelPosition="left" />
      <Group gap="md">
        <Text size="sm">EQ-2026-00042</Text>
        <Divider orientation="vertical" />
        <Text size="sm">En la Secretaría</Text>
      </Group>
    </Stack>
  ),
  parameters: {
    docs: {
      description: { story: 'Separa secciones de un mismo panel. Entre paneles alcanza con el espacio.' },
    },
  },
}

export const AspectRatioStory: StoryObj = {
  name: 'AspectRatio',
  render: () => (
    <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="md">
      <AspectRatio ratio={210 / 297}>
        <Center bg="fiuba.0">
          <Text size="sm">Vista previa del programa de la asignatura</Text>
        </Center>
      </AspectRatio>
    </SimpleGrid>
  ),
  parameters: {
    docs: {
      description: { story: 'Mantiene la proporción de una vista previa, como la de un documento en tamaño A4.' },
    },
  },
}

export const ScrollAreaStory: StoryObj = {
  name: 'ScrollArea',
  render: () => (
    <Paper withBorder>
      <AspectRatio ratio={3}>
        <ScrollArea type="auto" offsetScrollbars>
          <Stack gap="xs" p="md">
            {Array.from({ length: 12 }, (_, index) => (
              <Text key={index} size="sm">
                Movimiento {index + 1}: el trámite pasó al paso siguiente.
              </Text>
            ))}
          </Stack>
        </ScrollArea>
      </AspectRatio>
    </Paper>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Desplaza un contenido largo dentro de un espacio fijo, como el historial de movimientos de un trámite.',
      },
    },
  },
}

export const SplitterStory: StoryObj = {
  name: 'Splitter',
  render: () => (
    <Paper withBorder>
      <AspectRatio ratio={3}>
        <Splitter>
          <Splitter.Pane defaultSize={40} min={20} p="md">
            <Text size="sm">Bandeja de trámites</Text>
          </Splitter.Pane>
          <Splitter.Pane defaultSize={60} min={30} p="md" bg="fiuba.0">
            <Text size="sm">Detalle del trámite seleccionado</Text>
          </Splitter.Pane>
        </Splitter>
      </AspectRatio>
    </Paper>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Dos paneles que la persona puede redimensionar, como la bandeja junto al detalle del trámite.',
      },
    },
  },
}

function CollapseDemo() {
  const [opened, { toggle }] = useDisclosure(false)
  return (
    <Stack gap="sm" align="flex-start">
      <Button variant="default" onClick={toggle} aria-expanded={opened}>
        {opened ? 'Ocultar requisitos' : 'Ver requisitos'}
      </Button>
      <Collapse expanded={opened}>
        <Item>Certificado analítico y programas de las asignaturas aprobadas.</Item>
      </Collapse>
    </Stack>
  )
}

export const CollapseStory: StoryObj = {
  name: 'Collapse',
  render: () => <CollapseDemo />,
  parameters: {
    docs: {
      description: { story: 'Muestra y oculta un bloque secundario, como los requisitos de un tipo de trámite.' },
    },
  },
}

export const BoxStory: StoryObj = {
  name: 'Box',
  render: () => (
    <Box bg="fiuba.0" p="md">
      <Text size="sm">Un contenedor sin aspecto propio, que acepta las propiedades de estilo del tema.</Text>
    </Box>
  ),
  parameters: {
    docs: {
      description: { story: 'Para componentes propios. En una pantalla casi siempre hay un componente más preciso.' },
    },
  },
}

export const ScrollerStory: StoryObj = {
  name: 'Scroller',
  render: () => (
    <Scroller>
      <Group gap="xs" wrap="nowrap">
        {[
          'Equivalencia entre asignaturas',
          'Reconocimiento de créditos',
          'Pase de carrera',
          'Pase desde otra universidad',
          'Intercambio estudiantil',
          'Equivalencia parcial',
          'Reconocimiento de actividades',
        ].map((label) => (
          <Badge key={label} variant="light" size="lg">
            {label}
          </Badge>
        ))}
      </Group>
    </Scroller>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Una fila que se desplaza de costado con flechas, como los filtros por tipo de trámite en el celular.',
      },
    },
  },
}

export const AppShellStory: StoryObj = {
  name: 'AppShell',
  render: () => (
    <AspectRatio ratio={16 / 7}>
      <AppShell mode="static" header={{ height: 56 }} navbar={{ width: 220, breakpoint: 'xs' }} padding="md">
        <AppShell.Header p="md">
          <Text fw={700}>SIGTA</Text>
        </AppShell.Header>
        <AppShell.Navbar p="sm" aria-label="Menú principal">
          <Stack gap="xxs">
            <NavLink variant="menu" label="Inicio" leftSection={<Icon name="home" />} />
            <NavLink variant="menu" label="Bandeja de trámites" leftSection={<Icon name="inbox" />} active />
            <NavLink variant="menu" label="Catálogo de trámites" leftSection={<Icon name="catalog" />} />
          </Stack>
        </AppShell.Navbar>
        <AppShell.Main bg="gray.0">
          <Item>El contenido de la página.</Item>
        </AppShell.Main>
      </AppShell>
    </AspectRatio>
  ),
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        story:
          'El marco de las páginas con sesión: encabezado, menú lateral y contenido. Acá se muestra dentro de un recuadro, con `mode="static"`; el marco real de SIGTA está en SIGTA/Marco.',
      },
    },
  },
}
