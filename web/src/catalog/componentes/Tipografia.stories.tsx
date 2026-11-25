import { Blockquote, Code, Highlight, Mark, Stack, Text, Title, Typography } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'

// Text, headings and the components that mark a part of a text. Sizes, weights
// and colors come from the theme; the controls of the first story try them.

const meta: Meta = {
  title: 'Componentes/Tipografía',
  parameters: {
    docs: {
      description: {
        component:
          'El texto usa la tipografía del sistema y los tamaños `xs` a `xl` de `font.size` en `tokens.json`. Los títulos bajan de nivel sin saltear y cada página tiene un solo `Title order={1}`.',
      },
    },
  },
}

export default meta

export const TextStory: StoryObj<typeof Text> = {
  name: 'Text',
  render: (args) => <Text {...args} />,
  args: {
    children:
      'La solicitud de equivalencia entre asignaturas se evalúa con el programa de la asignatura aprobada y el certificado analítico de la carrera de origen.',
    size: 'md',
    fw: 400,
    c: undefined,
    truncate: false,
  },
  argTypes: {
    size: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
    fw: { control: 'select', options: [400, 500, 600] },
    c: { control: 'select', options: [undefined, 'dimmed', 'bright', 'fiuba', 'gray'] },
    truncate: { control: 'select', options: [false, true, 'start', 'end'] },
  },
  parameters: {
    docs: {
      description: {
        story:
          'Todo el texto de las pantallas. `dimmed` para lo secundario, como fechas y aclaraciones; `truncate` para nombres largos en listas y tablas.',
      },
    },
  },
}

export const TitleStory: StoryObj<typeof Title> = {
  name: 'Title',
  render: (args) => <Title {...args} />,
  args: { children: 'Reconocimiento de créditos', order: 1 },
  argTypes: {
    order: { control: 'select', options: [1, 2, 3, 4, 5, 6] },
    variant: { control: 'select', options: ['default', 'display', 'section'] },
  },
  parameters: {
    docs: {
      description: {
        story: 'El título de la página (`order={1}`) y de cada sección. Probá los niveles en los controles.',
      },
    },
  },
}

export const TitleOrders: StoryObj = {
  name: 'Title, niveles',
  render: () => (
    <Stack gap="sm">
      <Title variant="display">Los trámites académicos, en un solo lugar</Title>
      <Title order={1}>Bandeja de trámites</Title>
      <Title order={2}>Equivalencia entre asignaturas</Title>
      <Title order={3}>Documentación presentada</Title>
      <Title order={4}>Programa de la asignatura</Title>
      <Title order={5}>Observaciones de la comisión</Title>
      <Title order={6}>Historial de movimientos</Title>
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story:
          '`variant="display"` es el título de cada pantalla y el titular de la portada pública, y crece con el ancho de pantalla.',
      },
    },
  },
}

export const LandingType: StoryObj = {
  name: 'Portada',
  render: () => (
    <Stack gap="md">
      <Title variant="display">Los trámites académicos, en línea y con seguimiento</Title>
      <Title order={2} variant="section">
        Un solo lugar, del inicio a la resolución
      </Title>
      <Text variant="figure" c="fiuba.7">
        12
      </Text>
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'La escala grande: el titular (`Title variant="display"`), sin palabras destacadas ni etiquetas en mayúsculas encima, los títulos de sección (`Title variant="section"`) y las cifras (`Text variant="figure"`). Los tres tamaños crecen con el ancho de pantalla, entre los valores de `font.fluid` en `tokens.json`.',
      },
    },
  },
}

export const CodeStory: StoryObj = {
  name: 'Code',
  render: () => (
    <Stack gap="xs">
      <Text size="sm">
        Número de trámite: <Code>EQ-2026-00042</Code>
      </Text>
      <Code block>{'Código de asignatura: 61.03\nPlan de estudios: 2020'}</Code>
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Números de trámite, códigos de asignatura y otros identificadores que se copian tal cual.',
      },
    },
  },
}

export const HighlightStory: StoryObj = {
  name: 'Highlight',
  render: () => (
    <Highlight highlight={['equivalencia', 'Análisis Matemático']} size="sm">
      Solicitud de equivalencia de Análisis Matemático II por la asignatura aprobada en la carrera de origen.
    </Highlight>
  ),
  parameters: {
    docs: {
      description: { story: 'Resalta en los resultados de una búsqueda las palabras que se buscaron.' },
    },
  },
}

export const MarkStory: StoryObj = {
  name: 'Mark',
  render: () => (
    <Text size="sm">
      El programa de la asignatura tiene que estar <Mark>firmado por la unidad académica de origen</Mark>.
    </Text>
  ),
  parameters: {
    docs: {
      description: { story: 'Marca una parte de un texto que la persona no puede pasar por alto.' },
    },
  },
}

export const BlockquoteStory: StoryObj = {
  name: 'Blockquote',
  render: () => (
    <Blockquote cite="Comisión de equivalencias, dictamen de ejemplo">
      Los contenidos de la asignatura aprobada cubren los del programa vigente, por lo que corresponde otorgar la
      equivalencia total.
    </Blockquote>
  ),
  parameters: {
    docs: {
      description: { story: 'Cita textual de un dictamen o de una observación dentro del detalle de un trámite.' },
    },
  },
}

export const TypographyStory: StoryObj = {
  name: 'Typography',
  render: () => (
    <Typography>
      <h3>Requisitos del pase de carrera</h3>
      <p>
        Para solicitar el pase de carrera hay que tener aprobadas al menos dos asignaturas de la carrera de origen y
        presentar la siguiente documentación:
      </p>
      <ul>
        <li>Certificado analítico actualizado.</li>
        <li>Programas de las asignaturas aprobadas.</li>
        <li>Nota de solicitud dirigida a la Secretaría.</li>
      </ul>
    </Typography>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Da estilo a un texto con formato que no se arma con componentes, como los requisitos de un tipo de trámite.',
      },
    },
  },
}
