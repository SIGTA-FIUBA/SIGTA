import {
  AngleSlider,
  Checkbox,
  ColorInput,
  ColorPicker,
  Fieldset,
  FileInput,
  Group,
  Input,
  JsonInput,
  MaskInput,
  NumberInput,
  PasswordInput,
  PinInput,
  Radio,
  RangeSlider,
  Rating,
  SegmentedControl,
  Slider,
  Stack,
  Switch,
  Text,
  TextInput,
  Textarea,
  useMantineTheme,
} from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { useState } from 'react'

// Form fields: text, numbers, files, choices and ranges. Each story shows the
// component as SIGTA uses it; the controls of the first one try every property.

const meta: Meta = {
  title: 'Componentes/Entradas',
  parameters: {
    docs: {
      description: {
        component:
          'Campos de los formularios de trámite. Todo campo lleva `label`; `description` explica qué se espera y `error` dice qué hacer para corregirlo. El tamaño y el radio salen del tema.',
      },
    },
  },
}

export default meta

const subjects = [
  { value: 'am2', label: 'Análisis Matemático II', detail: 'Universidad Nacional de La Plata, 8 créditos' },
  { value: 'alg2', label: 'Álgebra II', detail: 'Universidad Nacional de La Plata, 6 créditos' },
  { value: 'fis1', label: 'Física I', detail: 'Universidad Tecnológica Nacional, 8 créditos' },
]

export const TextInputStory: StoryObj<typeof TextInput> = {
  name: 'TextInput',
  render: (args) => <TextInput {...args} />,
  args: {
    label: 'Número de padrón',
    description: 'El número que figura en la libreta universitaria',
    placeholder: '112233',
    error: '',
    required: false,
    disabled: false,
    size: 'sm',
  },
  argTypes: {
    size: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
    radius: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
    variant: { control: 'select', options: ['default', 'filled', 'unstyled'] },
  },
  parameters: {
    docs: {
      description: { story: 'El campo de texto de una línea. Probá las propiedades en los controles.' },
    },
  },
}

export const TextInputStates: StoryObj = {
  name: 'TextInput, estados',
  render: () => (
    <Stack>
      <TextInput label="Asignatura de origen" placeholder="Análisis Matemático II" />
      <TextInput
        label="Universidad de origen"
        description="Tal como figura en el certificado analítico"
        placeholder="Universidad Nacional de La Plata"
      />
      <TextInput label="Número de padrón" defaultValue="11A22" error="Ingresá un número de padrón válido" />
      <TextInput label="Correo electrónico institucional" placeholder="nombre@fi.uba.ar" required />
      <TextInput label="Carrera actual" defaultValue="Ingeniería en Informática" disabled />
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Etiqueta, descripción, error, obligatorio y deshabilitado. El error reemplaza a la descripción en rojo y dice qué hacer.',
      },
    },
  },
}

export const TextareaStory: StoryObj = {
  name: 'Textarea',
  render: () => (
    <Stack>
      <Textarea
        label="Motivo de la solicitud"
        description="Contá por qué pedís la equivalencia"
        placeholder="Cursé y aprobé la asignatura en otra universidad"
        rows={4}
      />
      <Textarea label="Observaciones de la Secretaría" error="Escribí al menos una observación" required />
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Texto libre de varias líneas: el motivo de un trámite o una observación de la Secretaría.',
      },
    },
  },
}

export const PasswordInputStory: StoryObj = {
  name: 'PasswordInput',
  render: () => (
    <PasswordInput
      label="Clave de acceso"
      description="Al menos 12 caracteres"
      visibilityToggleButtonProps={{ 'aria-label': 'Mostrar u ocultar la clave' }}
    />
  ),
  parameters: {
    docs: {
      description: {
        story: 'SIGTA no lo usa: el ingreso es con la cuenta de Google. Queda por si una integración pide una clave.',
      },
    },
  },
}

export const NumberInputStory: StoryObj = {
  name: 'NumberInput',
  render: () => (
    <Stack>
      <NumberInput
        label="Créditos a reconocer"
        description="Entre 1 y 20 créditos"
        min={1}
        max={20}
        defaultValue={8}
        allowDecimal={false}
      />
      <NumberInput
        label="Nota final"
        min={4}
        max={10}
        decimalScale={2}
        decimalSeparator=","
        defaultValue={11}
        error="Ingresá una nota entre 4 y 10"
      />
    </Stack>
  ),
  parameters: {
    docs: {
      description: { story: 'Cantidades con límites: créditos, carga horaria o la nota de una asignatura aprobada.' },
    },
  },
}

export const PinInputStory: StoryObj = {
  name: 'PinInput',
  render: () => (
    <Stack gap="xs">
      <Text size="sm">Código de verificación</Text>
      <PinInput length={6} type="number" oneTimeCode ariaLabel="Código de verificación" />
      <PinInput length={6} type="number" defaultValue="123" error ariaLabel="Código de verificación con error" />
      <Text size="xs" c="red">
        Ingresá los seis dígitos del código
      </Text>
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Un código corto, un dígito por casilla. Para confirmar una acción con un código enviado por correo, si se agrega.',
      },
    },
  },
}

export const MaskInputStory: StoryObj = {
  name: 'MaskInput',
  render: () => (
    <Stack>
      <MaskInput label="Número de documento" mask="99.999.999" placeholder="12.345.678" />
      <MaskInput
        label="Número de trámite"
        description="Figura en el comprobante de inicio"
        mask="EQ-9999-99999"
        alwaysShowMask
      />
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Datos con formato fijo, como un documento o un número de trámite. `9` es un dígito y `a` una letra.',
      },
    },
  },
}

export const JsonInputStory: StoryObj = {
  name: 'JsonInput',
  render: () => (
    <JsonInput
      label="Configuración del tipo de trámite"
      description="Se valida al salir del campo"
      validationError="Revisá el formato: no es un JSON válido"
      formatOnBlur
      rows={5}
      defaultValue={'{"tipo": "equivalencia", "requiereEvaluacionExterna": true, "plazoDias": 30}'}
    />
  ),
  parameters: {
    docs: {
      description: {
        story: 'Configuración avanzada de un tipo de trámite, que edita solo la administración.',
      },
    },
  },
}

export const FileInputStory: StoryObj = {
  name: 'FileInput',
  render: () => (
    <Stack>
      <FileInput
        label="Programa de la asignatura de origen"
        description="Un PDF de hasta 10 MB, firmado por la universidad de origen"
        placeholder="Elegí un archivo"
        accept="application/pdf"
        clearable
        required
      />
      <FileInput
        label="Documentación respaldatoria"
        placeholder="Elegí uno o más archivos"
        multiple
        error="Adjuntá el certificado analítico"
      />
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Adjunta documentación a un trámite. Para un botón suelto que abre el selector, `FileButton`.',
      },
    },
  },
}

function ColorInputDemo() {
  const theme = useMantineTheme()
  return (
    <ColorInput
      label="Color de la etiqueta"
      description="Solo los tonos de la escala primaria"
      format="hex"
      defaultValue={theme.colors.fiuba[7]}
      swatches={[...theme.colors.fiuba]}
      disallowInput
    />
  )
}

export const ColorInputStory: StoryObj = {
  name: 'ColorInput',
  render: () => <ColorInputDemo />,
  parameters: {
    docs: {
      description: {
        story: 'SIGTA no lo usa en las pantallas de trabajo: los colores salen del tema. Queda como referencia.',
      },
    },
  },
}

function ColorPickerDemo() {
  const theme = useMantineTheme()
  const [value, setValue] = useState(theme.colors.fiuba[7])
  return (
    <Stack gap="xs">
      <ColorPicker
        format="hex"
        value={value}
        onChange={setValue}
        withPicker={false}
        swatches={[...theme.colors.fiuba]}
        swatchesPerRow={10}
      />
      <Text size="sm">Tono elegido: {value}</Text>
    </Stack>
  )
}

export const ColorPickerStory: StoryObj = {
  name: 'ColorPicker',
  render: () => <ColorPickerDemo />,
  parameters: {
    docs: {
      description: {
        story: 'SIGTA no lo usa: los colores salen del tema. Si hace falta elegir uno, solo de una paleta fija.',
      },
    },
  },
}

export const CheckboxStory: StoryObj = {
  name: 'Checkbox',
  render: () => (
    <Stack>
      <Checkbox label="Declaro que la información presentada es verdadera" />
      <Checkbox label="Adjunté el programa de cada asignatura" defaultChecked />
      <Checkbox label="Todas las asignaturas del ciclo básico" indeterminate />
      <Checkbox label="Acepto el reglamento de equivalencias" error="Aceptá el reglamento para continuar" />
      <Checkbox label="Trámite cerrado" description="No se puede modificar" disabled defaultChecked />
    </Stack>
  ),
  parameters: {
    docs: {
      description: { story: 'Una opción que se marca o no: una declaración o una condición del trámite.' },
    },
  },
}

export const CheckboxGroupStory: StoryObj = {
  name: 'Checkbox.Group',
  render: () => (
    <Checkbox.Group
      label="Asignaturas a reconocer"
      description="Marcá todas las que querés incluir en la solicitud"
      defaultValue={['am2']}
      withAsterisk
    >
      <Stack gap="xs" mt="xs">
        {subjects.map((subject) => (
          <Checkbox key={subject.value} value={subject.value} label={subject.label} />
        ))}
      </Stack>
    </Checkbox.Group>
  ),
  parameters: {
    docs: {
      description: { story: 'Varias opciones de una misma pregunta, con una sola etiqueta para el grupo.' },
    },
  },
}

export const CheckboxCardStory: StoryObj = {
  name: 'Checkbox.Card',
  render: () => (
    <Checkbox.Group label="Asignaturas a reconocer" defaultValue={['alg2']}>
      <Stack gap="xs" mt="xs">
        {subjects.map((subject) => (
          <Checkbox.Card key={subject.value} value={subject.value} p="md">
            <Group wrap="nowrap" align="flex-start">
              <Checkbox.Indicator />
              <Stack gap={0}>
                <Text size="sm">{subject.label}</Text>
                <Text size="xs" c="dimmed">
                  {subject.detail}
                </Text>
              </Stack>
            </Group>
          </Checkbox.Card>
        ))}
      </Stack>
    </Checkbox.Group>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Cuando cada opción necesita más de una línea para elegirse bien, como una asignatura con su origen.',
      },
    },
  },
}

export const RadioStory: StoryObj = {
  name: 'Radio',
  render: () => (
    <Stack>
      <Radio label="Equivalencia entre asignaturas" name="procedure-states" defaultChecked />
      <Radio label="Reconocimiento de créditos" name="procedure-states" />
      <Radio label="Pase de carrera" name="procedure-states" description="Fuera del período de inscripción" disabled />
    </Stack>
  ),
  parameters: {
    docs: {
      description: { story: 'Una opción entre varias excluyentes. Casi siempre va dentro de `Radio.Group`.' },
    },
  },
}

export const RadioGroupStory: StoryObj = {
  name: 'Radio.Group',
  render: () => (
    <Stack>
      <Radio.Group label="Tipo de trámite" description="Elegí uno" defaultValue="equivalence" withAsterisk>
        <Stack gap="xs" mt="xs">
          <Radio value="equivalence" label="Equivalencia entre asignaturas" />
          <Radio value="credits" label="Reconocimiento de créditos" />
          <Radio value="transfer" label="Pase de carrera" />
        </Stack>
      </Radio.Group>
      <Radio.Group label="Modalidad de cursada en origen" error="Elegí una modalidad">
        <Group mt="xs">
          <Radio value="in-person" label="Presencial" />
          <Radio value="remote" label="A distancia" />
        </Group>
      </Radio.Group>
    </Stack>
  ),
  parameters: {
    docs: {
      description: { story: 'Una sola respuesta entre pocas opciones visibles. Con más de cinco, `Select`.' },
    },
  },
}

const procedureTypes = [
  {
    value: 'equivalence',
    label: 'Equivalencia entre asignaturas',
    detail: 'Por una asignatura aprobada en otra carrera',
  },
  { value: 'credits', label: 'Reconocimiento de créditos', detail: 'Por actividades académicas o extensión' },
  { value: 'transfer', label: 'Pase de carrera', detail: 'De otra carrera de la UBA a la FIUBA' },
]

export const RadioCardStory: StoryObj = {
  name: 'Radio.Card',
  render: () => (
    <Radio.Group label="Tipo de trámite" defaultValue="equivalence">
      <Stack gap="xs" mt="xs">
        {procedureTypes.map((type) => (
          <Radio.Card key={type.value} value={type.value} p="md">
            <Group wrap="nowrap" align="flex-start">
              <Radio.Indicator />
              <Stack gap={0}>
                <Text size="sm">{type.label}</Text>
                <Text size="xs" c="dimmed">
                  {type.detail}
                </Text>
              </Stack>
            </Group>
          </Radio.Card>
        ))}
      </Stack>
    </Radio.Group>
  ),
  parameters: {
    docs: {
      description: { story: 'Para elegir el tipo de trámite al iniciarlo, con una línea que explica cada uno.' },
    },
  },
}

export const SwitchStory: StoryObj = {
  name: 'Switch',
  render: () => (
    <Stack>
      <Switch label="Recibir avisos por correo" defaultChecked />
      <Switch label="Mostrar solo trámites con acción requerida" />
      <Switch label="Trámite visible para el titular" description="Lo define la Secretaría" disabled />
    </Stack>
  ),
  parameters: {
    docs: {
      description: { story: 'Una preferencia que se aplica al momento, sin botón de guardar.' },
    },
  },
}

export const SwitchGroupStory: StoryObj = {
  name: 'Switch.Group',
  render: () => (
    <Switch.Group label="Avisos por correo" description="Elegí cuándo recibirlos" defaultValue={['action']}>
      <Stack gap="xs" mt="xs">
        <Switch value="action" label="Cuando un trámite requiere una acción" />
        <Switch value="status" label="Cuando un trámite cambia de estado" />
        <Switch value="closed" label="Cuando un trámite se cierra" />
      </Stack>
    </Switch.Group>
  ),
  parameters: {
    docs: {
      description: { story: 'Varias preferencias relacionadas bajo una misma etiqueta.' },
    },
  },
}

export const SegmentedControlStory: StoryObj = {
  name: 'SegmentedControl',
  render: () => (
    <SegmentedControl
      aria-label="Filtrar trámites por estado"
      data={[
        { value: 'all', label: 'Todos' },
        { value: 'in_office', label: 'En un área' },
        { value: 'referred', label: 'Derivado' },
        { value: 'with_applicant', label: 'Pendiente del titular' },
      ]}
    />
  ),
  parameters: {
    docs: {
      description: { story: 'Cambia entre vistas de una misma lista, como los estados de la bandeja.' },
    },
  },
}

export const SliderStory: StoryObj = {
  name: 'Slider',
  render: () => (
    <Stack gap="xs">
      <Text size="sm">Coincidencia de contenidos con el programa de la FIUBA</Text>
      <Slider
        aria-label="Coincidencia de contenidos con el programa de la FIUBA"
        defaultValue={75}
        step={5}
        label={(value) => `${value} %`}
        marks={[
          { value: 0, label: '0 %' },
          { value: 50, label: '50 %' },
          { value: 100, label: '100 %' },
        ]}
        mb="lg"
      />
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Un valor aproximado dentro de un rango, como el grado de coincidencia que estima quien evalúa.',
      },
    },
  },
}

export const RangeSliderStory: StoryObj = {
  name: 'RangeSlider',
  render: () => (
    <Stack gap="xs">
      <Text size="sm">Año de aprobación</Text>
      <RangeSlider
        aria-label="Año de aprobación"
        min={2010}
        max={2026}
        minRange={1}
        defaultValue={[2018, 2024]}
        marks={[
          { value: 2010, label: '2010' },
          { value: 2018, label: '2018' },
          { value: 2026, label: '2026' },
        ]}
        mb="lg"
      />
    </Stack>
  ),
  parameters: {
    docs: {
      description: { story: 'Un filtro entre dos valores, como el período en que se aprobaron las asignaturas.' },
    },
  },
}

export const AngleSliderStory: StoryObj = {
  name: 'AngleSlider',
  render: () => (
    <Group>
      <AngleSlider
        aria-label="Rotación del documento"
        defaultValue={90}
        step={90}
        formatLabel={(value) => `${value}°`}
      />
      <AngleSlider aria-label="Rotación deshabilitada" defaultValue={0} disabled />
    </Group>
  ),
  parameters: {
    docs: {
      description: {
        story: 'SIGTA no lo usa. Serviría para rotar la vista previa de un documento escaneado.',
      },
    },
  },
}

export const RatingStory: StoryObj = {
  name: 'Rating',
  render: () => (
    <Stack gap="xs">
      <Text size="sm">Valoración de la atención recibida</Text>
      <Rating defaultValue={4} getSymbolLabel={(index) => `${index + 1} de 5`} />
      <Rating value={3} readOnly />
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Para una encuesta de satisfacción al cerrar un trámite, si se agrega. `readOnly` muestra un promedio.',
      },
    },
  },
}

export const FieldsetStory: StoryObj = {
  name: 'Fieldset',
  render: () => (
    <Stack>
      <Fieldset legend="Asignatura de origen">
        <Stack>
          <TextInput label="Nombre de la asignatura" placeholder="Análisis Matemático II" required />
          <TextInput label="Universidad" placeholder="Universidad Nacional de La Plata" required />
          <NumberInput label="Carga horaria total" description="En horas" min={1} />
        </Stack>
      </Fieldset>
      <Fieldset legend="Asignatura de la FIUBA" disabled>
        <Stack>
          <TextInput label="Nombre de la asignatura" defaultValue="Análisis Matemático II" />
          <TextInput label="Carrera" defaultValue="Ingeniería en Informática" />
        </Stack>
      </Fieldset>
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Agrupa los campos de una parte del formulario, como cada asignatura de una equivalencia. `disabled` bloquea todos sus campos.',
      },
    },
  },
}

export const InputStory: StoryObj = {
  name: 'Input',
  render: () => (
    <Stack>
      <Input.Wrapper
        id="presentation-date"
        label="Fecha de presentación"
        description="Día en que se presentó la solicitud en la Secretaría"
        required
      >
        <Input id="presentation-date" placeholder="14/11/2026" />
      </Input.Wrapper>
      <Input.Wrapper id="file-number" label="Número de expediente" error="Ingresá el número completo del expediente">
        <Input id="file-number" defaultValue="EXP-2026" error />
      </Input.Wrapper>
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'La base de los demás campos. `Input.Wrapper` agrega etiqueta, descripción y error a un control propio; para un campo común, `TextInput`.',
      },
    },
  },
}
