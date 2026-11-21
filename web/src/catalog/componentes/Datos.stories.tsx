import {
  Accordion,
  Avatar,
  BackgroundImage,
  Badge,
  Button,
  Card,
  ColorSwatch,
  DataList,
  Group,
  Image,
  Kbd,
  List,
  Marquee,
  NumberFormatter,
  OverflowList,
  Paper,
  RollingNumber,
  SimpleGrid,
  Spoiler,
  Stack,
  Table,
  Text,
  ThemeIcon,
  Timeline,
} from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { useState } from 'react'
import { MenuIcon } from '../../layout/MenuIcon'

// Components that show data: tables, states, people, panels and lists. Each
// story shows the component as SIGTA uses it; the controls of the first ones
// try every property.

const meta: Meta = {
  title: 'Componentes/Datos',
  parameters: {
    docs: {
      description: {
        component:
          'Los datos se muestran en paneles `Paper withBorder` y tablas con `highlightOnHover`. Los estados de un trámite tienen color fijo: "En un área" en el primario, "Derivado" en naranja y "Pendiente del titular" en gris.',
      },
    },
  },
}

export default meta

const procedures = [
  {
    id: 'EQ-2026-00042',
    applicant: 'Lucía Benítez',
    type: 'Equivalencia entre asignaturas',
    step: 'Revisión de la documentación',
    area: 'Departamento de Matemática',
    state: 'En un área',
    color: 'fiuba',
    age: '12 días',
  },
  {
    id: 'RC-2026-00017',
    applicant: 'Tomás Aguirre',
    type: 'Reconocimiento de créditos',
    step: 'Dictamen del evaluador',
    area: 'Comisión curricular',
    state: 'Derivado',
    color: 'orange',
    age: '30 días',
  },
  {
    id: 'PC-2026-00008',
    applicant: 'Martina Sosa',
    type: 'Pase de carrera',
    step: 'Presentación del certificado analítico',
    area: 'Titular del trámite',
    state: 'Pendiente del titular',
    color: 'gray',
    age: '5 días',
  },
]

export const TableStory: StoryObj = {
  name: 'Table',
  render: () => (
    <Table.ScrollContainer minWidth={640}>
      <Table verticalSpacing="sm" highlightOnHover>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Trámite</Table.Th>
            <Table.Th>Instancia pendiente</Table.Th>
            <Table.Th>Estado</Table.Th>
            <Table.Th ta="right">Antigüedad</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {procedures.map((item) => (
            <Table.Tr key={item.id}>
              <Table.Td>
                <Stack gap={0}>
                  <Text fw={600}>{item.applicant}</Text>
                  <Text size="sm" c="dimmed">
                    {item.type}
                  </Text>
                </Stack>
              </Table.Td>
              <Table.Td>
                <Stack gap={0}>
                  <Text fw={500}>{item.step}</Text>
                  <Text size="sm" c="dimmed">
                    {item.area}
                  </Text>
                </Stack>
              </Table.Td>
              <Table.Td>
                <Badge variant="light" color={item.color}>
                  {item.state}
                </Badge>
              </Table.Td>
              <Table.Td ta="right">{item.age}</Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </Table.ScrollContainer>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'La lista de trámites de la bandeja. Las filas se resaltan al pasar el mouse y `Table.ScrollContainer` permite desplazarla en pantallas angostas.',
      },
    },
  },
}

export const BadgeStory: StoryObj<typeof Badge> = {
  name: 'Badge',
  render: (args) => <Badge {...args} />,
  args: { children: 'Acción requerida', variant: 'filled', size: 'sm' },
  argTypes: {
    variant: { control: 'select', options: ['filled', 'light', 'outline', 'dot', 'default', 'transparent', 'white'] },
    size: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
    color: { control: 'select', options: ['fiuba', 'orange', 'gray'] },
    radius: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
  },
  parameters: {
    docs: {
      description: {
        story: 'Una etiqueta corta, como "Acción requerida" en la bandeja. Probá las propiedades en los controles.',
      },
    },
  },
}

export const BadgeStates: StoryObj = {
  name: 'Badge, estados',
  render: () => (
    <Group gap="xs">
      <Badge variant="light" color="fiuba">
        En un área
      </Badge>
      <Badge variant="light" color="orange">
        Derivado
      </Badge>
      <Badge variant="light" color="gray">
        Pendiente del titular
      </Badge>
    </Group>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Los tres estados de un trámite, en `variant="light"`. Los colores son fijos y el rojo no se usa.',
      },
    },
  },
}

export const AvatarStory: StoryObj = {
  name: 'Avatar',
  render: () => (
    <Group>
      <Avatar name="Lucía Benítez" color="initials" />
      <Avatar name="Tomás Aguirre" color="initials" />
      <Avatar.Group>
        <Avatar name="Martina Sosa" color="initials" />
        <Avatar name="Julián Paredes" color="initials" />
        <Avatar name="Carla Domínguez" color="initials" />
        <Avatar>+2</Avatar>
      </Avatar.Group>
    </Group>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'La persona con la sesión, en el menú de la cuenta. Las iniciales salen de `name`, sin cambiar el nombre. `Avatar.Group` junta a quienes intervienen en un trámite.',
      },
    },
  },
}

export const CardStory: StoryObj = {
  name: 'Card',
  render: () => (
    <Card withBorder padding="lg" maw="50%">
      <Card.Section withBorder inheritPadding py="sm">
        <Text fw={600}>Equivalencia entre asignaturas</Text>
      </Card.Section>
      <Text size="sm" c="dimmed" mt="sm">
        Reconoce una asignatura aprobada en otra carrera o institución como equivalente a una del plan de estudios.
      </Text>
      <Group mt="md">
        <Button variant="light">Ver requisitos</Button>
      </Group>
    </Card>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Un panel con secciones separadas. Para un panel simple, SIGTA usa `Paper withBorder`.',
      },
    },
  },
}

export const PaperStory: StoryObj = {
  name: 'Paper',
  render: () => (
    <SimpleGrid cols={2} spacing="sm">
      <Paper withBorder p="md">
        <Stack gap="xs">
          <Text fw={600}>Análisis Matemático II</Text>
          <Text size="sm" c="dimmed">
            Asignatura de origen aprobada el 12 de marzo de 2025.
          </Text>
        </Stack>
      </Paper>
      <Paper component="button" type="button" variant="interactive" withBorder p="md" ta="left">
        <Stack gap="xs">
          <Text size="sm" fw={500} c="dimmed">
            Acción requerida
          </Text>
          <Text fz="xl" fw={700}>
            7
          </Text>
          <Text size="xs" c="dimmed">
            Instancias asignadas a este perfil
          </Text>
        </Stack>
      </Paper>
    </SimpleGrid>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'El panel de SIGTA es `Paper withBorder`. Si el panel es un control, como los resúmenes que filtran la bandeja, usa `variant="interactive"`: sube al pasar el mouse y se asienta al presionarlo.',
      },
    },
  },
}

export const TimelineStory: StoryObj = {
  name: 'Timeline',
  render: () => (
    <Timeline active={1} bulletSize={20} lineWidth={2}>
      <Timeline.Item title="Solicitud iniciada">
        <Text size="sm" c="dimmed">
          El titular presentó la equivalencia de Análisis Matemático II.
        </Text>
      </Timeline.Item>
      <Timeline.Item title="Revisión de la documentación">
        <Text size="sm" c="dimmed">
          En el Departamento de Matemática.
        </Text>
      </Timeline.Item>
      <Timeline.Item title="Evaluación externa" lineVariant="dashed">
        <Text size="sm" c="dimmed">
          Pendiente.
        </Text>
      </Timeline.Item>
      <Timeline.Item title="Resolución">
        <Text size="sm" c="dimmed">
          Pendiente.
        </Text>
      </Timeline.Item>
    </Timeline>
  ),
  parameters: {
    docs: {
      description: { story: 'El recorrido de un trámite: las instancias cumplidas, la actual y las que faltan.' },
    },
  },
}

export const ListStory: StoryObj = {
  name: 'List',
  render: () => (
    <List size="sm" spacing="xs">
      <List.Item>Certificado analítico de la carrera de origen</List.Item>
      <List.Item>Programa de la asignatura aprobada</List.Item>
      <List.Item>Plan de estudios vigente al momento de la aprobación</List.Item>
    </List>
  ),
  parameters: {
    docs: { description: { story: 'Requisitos y documentación de un trámite en el catálogo.' } },
  },
}

export const DataListStory: StoryObj = {
  name: 'DataList',
  render: () => (
    <DataList withDivider>
      <DataList.Item>
        <DataList.ItemLabel>Número de trámite</DataList.ItemLabel>
        <DataList.ItemValue>EQ-2026-00042</DataList.ItemValue>
      </DataList.Item>
      <DataList.Item>
        <DataList.ItemLabel>Tipo</DataList.ItemLabel>
        <DataList.ItemValue>Equivalencia entre asignaturas</DataList.ItemValue>
      </DataList.Item>
      <DataList.Item>
        <DataList.ItemLabel>Titular</DataList.ItemLabel>
        <DataList.ItemValue>Lucía Benítez</DataList.ItemValue>
      </DataList.Item>
      <DataList.Item>
        <DataList.ItemLabel>Estado</DataList.ItemLabel>
        <DataList.ItemValue>
          <Badge variant="light" color="fiuba">
            En un área
          </Badge>
        </DataList.ItemValue>
      </DataList.Item>
    </DataList>
  ),
  parameters: {
    docs: {
      description: { story: 'Los datos de un trámite como pares de etiqueta y valor, en el detalle del trámite.' },
    },
  },
}

export const AccordionStory: StoryObj = {
  name: 'Accordion',
  render: () => (
    <Accordion variant="separated" defaultValue="requisitos">
      <Accordion.Item value="requisitos">
        <Accordion.Control>Requisitos</Accordion.Control>
        <Accordion.Panel>Tener aprobada la asignatura de origen con un programa equivalente.</Accordion.Panel>
      </Accordion.Item>
      <Accordion.Item value="plazos">
        <Accordion.Control>Plazos</Accordion.Control>
        <Accordion.Panel>La Secretaría resuelve dentro de los 60 días hábiles desde la presentación.</Accordion.Panel>
      </Accordion.Item>
      <Accordion.Item value="documentacion">
        <Accordion.Control>Documentación</Accordion.Control>
        <Accordion.Panel>Certificado analítico y programa de la asignatura, en PDF.</Accordion.Panel>
      </Accordion.Item>
    </Accordion>
  ),
  parameters: {
    docs: { description: { story: 'Secciones que se abren de a una, como el detalle de un trámite del catálogo.' } },
  },
}

export const SpoilerStory: StoryObj = {
  name: 'Spoiler',
  render: () => (
    <Spoiler maxHeight={48} showLabel="Ver más" hideLabel="Ver menos">
      <Text size="sm">
        Programa de Análisis Matemático II: funciones de varias variables, límite y continuidad, derivadas parciales,
        diferenciabilidad, extremos libres y condicionados, integrales múltiples, integrales de línea y de superficie,
        teoremas de Green, Gauss y Stokes, y ecuaciones diferenciales ordinarias de primer y segundo orden.
      </Text>
    </Spoiler>
  ),
  parameters: {
    docs: { description: { story: 'Recorta un texto largo, como el programa de una asignatura, hasta que se pida.' } },
  },
}

export const ImageStory: StoryObj = {
  name: 'Image',
  render: () => <Image src="brand/sigta-isotipo-color.svg" alt="Isotipo de SIGTA" h="xl" w="auto" fit="contain" />,
  parameters: {
    docs: {
      description: {
        story: 'Solo para la marca: el logo del encabezado. La interfaz no usa fotos ni imágenes decorativas.',
      },
    },
  },
}

export const BackgroundImageStory: StoryObj = {
  name: 'BackgroundImage',
  render: () => (
    <BackgroundImage src="brand/sigta-isotipo-color.svg" p="xl">
      <Text fw={600}>Sistema Integral de Gestión de Trámites Académicos</Text>
    </BackgroundImage>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Parte de Mantine; SIGTA no lo usa, porque las pantallas de trabajo no llevan fondos decorativos.',
      },
    },
  },
}

export const ColorSwatchStory: StoryObj = {
  name: 'ColorSwatch',
  render: () => (
    <Group gap="lg">
      {[
        { color: 'var(--mantine-color-fiuba-7)', label: 'En un área' },
        { color: 'var(--mantine-color-orange-6)', label: 'Derivado' },
        { color: 'var(--mantine-color-gray-6)', label: 'Pendiente del titular' },
      ].map((swatch) => (
        <Group key={swatch.label} gap="xs">
          <ColorSwatch color={swatch.color} />
          <Text size="sm">{swatch.label}</Text>
        </Group>
      ))}
    </Group>
  ),
  parameters: {
    docs: { description: { story: 'Muestra un color, por ejemplo en la leyenda de un gráfico por estado.' } },
  },
}

export const ThemeIconStory: StoryObj = {
  name: 'ThemeIcon',
  render: () => (
    <Group>
      {[1, 2, 3].map((step) => (
        <ThemeIcon key={step} variant="light" radius="xl" aria-hidden>
          <Text size="sm" fw={700}>
            {step}
          </Text>
        </ThemeIcon>
      ))}
      <ThemeIcon variant="light" size="lg" aria-hidden>
        <MenuIcon name="catalog" />
      </ThemeIcon>
    </Group>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Un ícono o un número sobre un fondo de color, como los pasos de "Cómo se realiza un trámite" en el inicio.',
      },
    },
  },
}

export const KbdStory: StoryObj = {
  name: 'Kbd',
  render: () => (
    <Text size="sm">
      Para buscar un trámite, presioná <Kbd>Ctrl</Kbd> + <Kbd>K</Kbd>.
    </Text>
  ),
  parameters: {
    docs: { description: { story: 'Una tecla en un texto de ayuda, para nombrar un atajo de teclado.' } },
  },
}

export const NumberFormatterStory: StoryObj = {
  name: 'NumberFormatter',
  render: () => (
    <Stack gap="xs">
      <Text size="sm">
        Carga horaria reconocida: <NumberFormatter value={1280} thousandSeparator="." decimalSeparator="," /> horas
      </Text>
      <Text size="sm">
        Promedio: <NumberFormatter value={7.5} decimalScale={2} fixedDecimalScale decimalSeparator="," />
      </Text>
    </Stack>
  ),
  parameters: {
    docs: {
      description: { story: 'Un número con el formato local: punto para los miles y coma para los decimales.' },
    },
  },
}

function RollingNumberDemo() {
  const [count, setCount] = useState(7)
  return (
    <Group>
      <RollingNumber value={count} fz="xl" fw={700} />
      <Button variant="default" onClick={() => setCount((value) => value + 1)}>
        Sumar un trámite
      </Button>
    </Group>
  )
}

export const RollingNumberStory: StoryObj = {
  name: 'RollingNumber',
  render: () => <RollingNumberDemo />,
  parameters: {
    docs: {
      description: {
        story: 'Un número que cambia con una transición de dígitos, como un contador de trámites. Usar con moderación.',
      },
    },
  },
}

export const MarqueeStory: StoryObj = {
  name: 'Marquee',
  render: () => (
    <Marquee pauseOnHover>
      <Text size="sm">Equivalencia entre asignaturas</Text>
      <Text size="sm">Reconocimiento de créditos</Text>
      <Text size="sm">Pase de carrera</Text>
    </Marquee>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Parte de Mantine; SIGTA no lo usa, porque el movimiento continuo distrae en una herramienta de trabajo.',
      },
    },
  },
}

const subjects = [
  'Análisis Matemático II',
  'Álgebra Lineal',
  'Física I',
  'Química',
  'Probabilidad y Estadística',
  'Algoritmos y Programación I',
  'Física II',
  'Análisis Numérico',
]

export const OverflowListStory: StoryObj = {
  name: 'OverflowList',
  render: () => (
    <OverflowList
      data={subjects}
      maxVisibleItems={4}
      renderItem={(subject) => (
        <Badge key={subject} variant="default">
          {subject}
        </Badge>
      )}
      renderOverflow={(hidden) => <Badge variant="light">{`${hidden.length} más`}</Badge>}
    />
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Una fila de elementos que muestra los que entran y resume el resto, como las asignaturas de una solicitud.',
      },
    },
  },
}
