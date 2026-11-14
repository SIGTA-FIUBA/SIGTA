import {
  Anchor,
  Avatar,
  Badge,
  Box,
  Breadcrumbs,
  Container,
  getTreeExpandedState,
  Grid,
  Group,
  Menu,
  Menubar,
  NavLink,
  Pagination,
  Paper,
  Stack,
  Stepper,
  TableOfContents,
  Tabs,
  Text,
  Title,
  Tree,
  type TreeNodeData,
  useTree,
} from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { MenuIcon } from '../../components/MenuIcon/MenuIcon'

// Links and the controls that move between pages or parts of a page. Each story
// shows the component as SIGTA uses it.

const meta: Meta = {
  title: 'Componentes/Navegación',
  parameters: {
    docs: {
      description: {
        component:
          'El menú lateral es `NavLink variant="menu"` sobre el celeste de la marca; el resto de los links usan `Anchor`. Ninguna historia navega: los links apuntan a `#`.',
      },
    },
  },
}

export default meta

export const NavLinkStory: StoryObj<typeof NavLink> = {
  name: 'NavLink',
  render: (args) => <NavLink href="#" leftSection={<MenuIcon name="inbox" />} {...args} />,
  args: {
    label: 'Bandeja de trámites',
    description: 'Trámites con acción requerida',
    active: false,
    disabled: false,
    variant: 'light',
  },
  argTypes: {
    variant: { control: 'select', options: ['filled', 'light', 'subtle'] },
    color: { control: 'select', options: ['fiuba', 'gray'] },
  },
  parameters: {
    docs: {
      description: {
        story:
          'Un link de una lista de navegación, con ícono y descripción opcionales. Probá las propiedades en los controles.',
      },
    },
  },
}

export const NavLinkMenu: StoryObj = {
  name: 'NavLink, menú lateral',
  render: () => (
    <Container size="xs" mx={0} px={0}>
      <Box bg="fiuba.5" p="sm">
        <Stack gap="xs">
          <NavLink href="#" label="Inicio" leftSection={<MenuIcon name="home" />} variant="menu" fw={600} />
          <NavLink
            href="#"
            label="Bandeja de trámites"
            leftSection={<MenuIcon name="inbox" />}
            rightSection={
              <Badge variant="white" size="sm" circle aria-label="3 con acción requerida">
                3
              </Badge>
            }
            active
            variant="menu"
            fw={600}
          />
          <NavLink
            href="#"
            label="Catálogo de trámites"
            leftSection={<MenuIcon name="catalog" />}
            variant="menu"
            fw={600}
          />
          <Paper p="sm" radius="md" shadow="xs" mt="xl">
            <Group gap="sm" wrap="nowrap">
              <Avatar name="Laura Méndez" color="fiuba" aria-hidden />
              <Stack gap={0} miw={0}>
                <Text size="sm" fw={600} truncate="end">
                  Laura Méndez
                </Text>
                <Text size="xs" c="dimmed">
                  Perfil operativo
                </Text>
              </Stack>
            </Group>
          </Paper>
        </Stack>
      </Box>
    </Container>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'El menú de las pantallas con sesión (`AppLayout`): recuadros blancos sobre `fiuba.5` y la página actual rellena. Sobre el celeste no va texto suelto.',
      },
    },
  },
}

export const AnchorStory: StoryObj<typeof Anchor> = {
  name: 'Anchor',
  render: (args) => <Anchor href="#" {...args} />,
  args: { children: 'Ver el reglamento de equivalencias', size: 'sm', underline: 'hover' },
  argTypes: {
    underline: { control: 'select', options: ['always', 'hover', 'not-hover', 'never'] },
    size: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
  },
  parameters: {
    docs: {
      description: { story: 'El link dentro de un texto o una tabla, como el número de un trámite en la bandeja.' },
    },
  },
}

export const BreadcrumbsStory: StoryObj = {
  name: 'Breadcrumbs',
  render: () => (
    <Breadcrumbs>
      <Anchor href="#" size="sm">
        Inicio
      </Anchor>
      <Anchor href="#" size="sm">
        Bandeja de trámites
      </Anchor>
      <Text size="sm">EQ-2026-00042</Text>
    </Breadcrumbs>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Sobre el título de una página interna, como el detalle de un trámite, para volver a la lista.',
      },
    },
  },
}

const paginationLabels = {
  first: 'Primera página',
  previous: 'Página anterior',
  next: 'Página siguiente',
  last: 'Última página',
}

export const PaginationStory: StoryObj<typeof Pagination> = {
  name: 'Pagination',
  render: (args) => (
    <Pagination getControlProps={(control) => ({ 'aria-label': paginationLabels[control] })} {...args} />
  ),
  args: { total: 12, defaultValue: 1, siblings: 1, boundaries: 1, withEdges: false, size: 'sm' },
  argTypes: { size: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] } },
  parameters: {
    docs: {
      description: {
        story: 'Debajo de una lista larga, como la bandeja de trámites. Los controles sin texto llevan `aria-label`.',
      },
    },
  },
}

export const TabsStory: StoryObj = {
  name: 'Tabs',
  render: () => (
    <Tabs defaultValue="request">
      <Tabs.List>
        <Tabs.Tab value="request">Solicitud</Tabs.Tab>
        <Tabs.Tab value="documents">Documentación</Tabs.Tab>
        <Tabs.Tab value="history">Historial</Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="request" pt="md">
        <Text size="sm">Equivalencia entre asignaturas, presentada el 3 de noviembre.</Text>
      </Tabs.Panel>
      <Tabs.Panel value="documents" pt="md">
        <Text size="sm">Programa de la asignatura y certificado analítico.</Text>
      </Tabs.Panel>
      <Tabs.Panel value="history" pt="md">
        <Text size="sm">Enviada a la Secretaría. En revisión de la documentación.</Text>
      </Tabs.Panel>
    </Tabs>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Partes de una misma página que no se ven a la vez, como las secciones del detalle de un trámite.',
      },
    },
  },
}

export const StepperStory: StoryObj = {
  name: 'Stepper',
  render: () => (
    <Stepper active={2} size="sm">
      <Stepper.Step label="Solicitud" description="Presentada por el titular" />
      <Stepper.Step label="Revisión de la Secretaría" description="Control de la documentación" />
      <Stepper.Step label="Evaluación del departamento" description="Dictamen sobre la equivalencia" />
      <Stepper.Step label="Resolución" description="Decisión final" />
    </Stepper>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Las etapas de un trámite y en cuál está. Solo muestra el avance: las etapas no se eligen con un clic.',
      },
    },
  },
}

export const TableOfContentsStory: StoryObj = {
  name: 'TableOfContents',
  render: () => (
    <Grid>
      <Grid.Col span={{ base: 12, sm: 4 }}>
        <TableOfContents
          variant="light"
          size="sm"
          minDepthToOffset={3}
          scrollSpyOptions={{ selector: '[data-toc-procedure]' }}
          initialData={[
            { depth: 3, value: 'Requisitos' },
            { depth: 3, value: 'Documentación' },
            { depth: 4, value: 'Programa de la asignatura' },
            { depth: 3, value: 'Plazos' },
          ]}
          getControlProps={({ data }) => ({
            onClick: () => data.getNode().scrollIntoView({ behavior: 'smooth', block: 'center' }),
            children: data.value,
          })}
        />
      </Grid.Col>
      <Grid.Col span={{ base: 12, sm: 8 }}>
        <Stack gap="xs">
          <Title order={3} data-toc-procedure>
            Requisitos
          </Title>
          <Text size="sm">Haber aprobado la asignatura en otra carrera o universidad.</Text>
          <Title order={3} data-toc-procedure>
            Documentación
          </Title>
          <Text size="sm">Certificado analítico con las calificaciones.</Text>
          <Title order={4} data-toc-procedure>
            Programa de la asignatura
          </Title>
          <Text size="sm">El programa vigente cuando se cursó, con la carga horaria.</Text>
          <Title order={3} data-toc-procedure>
            Plazos
          </Title>
          <Text size="sm">La Secretaría revisa la solicitud dentro de los diez días hábiles.</Text>
        </Stack>
      </Grid.Col>
    </Grid>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'El índice de una página larga, como la de un tipo de trámite del catálogo. Solo lee los títulos marcados con `scrollSpyOptions.selector`.',
      },
    },
  },
}

export const MenubarStory: StoryObj = {
  name: 'Menubar',
  render: () => (
    <Menubar>
      <Menubar.Menu>
        <Menubar.Target>Trámite</Menubar.Target>
        <Menubar.Dropdown>
          <Menu.Item>Guardar borrador</Menu.Item>
          <Menu.Item>Descargar comprobante</Menu.Item>
          <Menu.Divider />
          <Menu.Item>Imprimir el resumen</Menu.Item>
        </Menubar.Dropdown>
      </Menubar.Menu>
      <Menubar.Menu>
        <Menubar.Target>Documentación</Menubar.Target>
        <Menubar.Dropdown>
          <Menu.Item>Adjuntar archivo</Menu.Item>
          <Menu.Item>Ver los adjuntos</Menu.Item>
        </Menubar.Dropdown>
      </Menubar.Menu>
      <Menubar.Menu>
        <Menubar.Target>Ayuda</Menubar.Target>
        <Menubar.Dropdown>
          <Menu.Item>Reglamento de equivalencias</Menu.Item>
          <Menu.Item>Preguntas frecuentes</Menu.Item>
        </Menubar.Dropdown>
      </Menubar.Menu>
    </Menubar>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Una barra de menús para una pantalla de trabajo con muchas acciones agrupadas. Con pocas acciones alcanza con botones o un `Menu`.',
      },
    },
  },
}

const studyPlan: TreeNodeData[] = [
  {
    label: 'Ciclo básico',
    value: 'basic',
    children: [
      { label: 'Análisis Matemático II', value: 'analysis-2' },
      { label: 'Álgebra Lineal', value: 'linear-algebra' },
      { label: 'Física I', value: 'physics-1' },
    ],
  },
  {
    label: 'Ciclo profesional',
    value: 'professional',
    children: [
      { label: 'Algoritmos y Estructuras de Datos', value: 'algorithms' },
      { label: 'Sistemas Operativos', value: 'operating-systems' },
    ],
  },
]

function TreeDemo() {
  const tree = useTree({ initialExpandedState: getTreeExpandedState(studyPlan, '*') })
  return <Tree data={studyPlan} tree={tree} selectOnClick withLines />
}

export const TreeStory: StoryObj = {
  name: 'Tree',
  render: () => <TreeDemo />,
  parameters: {
    docs: {
      description: {
        story:
          'Datos con jerarquía, como las asignaturas de un plan de estudios agrupadas por ciclo al elegir cuáles entran en una equivalencia.',
      },
    },
  },
}
