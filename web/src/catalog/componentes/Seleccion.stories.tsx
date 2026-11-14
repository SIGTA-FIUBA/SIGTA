import {
  Autocomplete,
  Cascader,
  Combobox,
  InputBase,
  MultiSelect,
  NativeSelect,
  Pill,
  PillsInput,
  Select,
  Stack,
  TagsInput,
  TreeSelect,
  useCombobox,
} from '@mantine/core'
import type { CascaderOption, TreeNodeData } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { useState } from 'react'

// Fields that pick one or more values from a list. Each story shows the
// component as SIGTA uses it; the controls of the first one try every property.

const meta: Meta = {
  title: 'Componentes/Selección',
  parameters: {
    docs: {
      description: {
        component:
          'Campos que eligen uno o más valores de una lista. Con pocas opciones visibles conviene `Radio.Group` o `Checkbox.Group`; con muchas, un campo con búsqueda. Todo campo lleva `label` y un mensaje para cuando la búsqueda no encuentra nada.',
      },
    },
  },
}

export default meta

const degrees = [
  'Ingeniería en Informática',
  'Ingeniería Civil',
  'Ingeniería Electrónica',
  'Ingeniería Industrial',
  'Ingeniería Mecánica',
  'Ingeniería Química',
  'Licenciatura en Análisis de Sistemas',
]

const subjects = [
  'Análisis Matemático II',
  'Álgebra II',
  'Física I',
  'Física II',
  'Probabilidad y Estadística',
  'Algoritmos y Estructuras de Datos',
  'Base de Datos',
  'Sistemas Operativos',
]

export const SelectStory: StoryObj<typeof Select> = {
  name: 'Select',
  render: (args) => <Select {...args} />,
  args: {
    label: 'Carrera de destino',
    description: 'La carrera de la FIUBA a la que pedís el pase',
    placeholder: 'Elegí una carrera',
    data: degrees,
    searchable: true,
    clearable: false,
    nothingFoundMessage: 'No hay carreras que coincidan',
    error: '',
    required: false,
    disabled: false,
    size: 'sm',
  },
  argTypes: {
    size: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
    radius: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
  },
  parameters: {
    docs: {
      description: {
        story: 'Una opción de una lista larga, como la carrera. Probá las propiedades en los controles.',
      },
    },
  },
}

export const SelectStates: StoryObj = {
  name: 'Select, estados',
  render: () => (
    <Stack>
      <Select
        label="Tipo de trámite"
        placeholder="Elegí un tipo de trámite"
        data={[
          {
            group: 'Equivalencias',
            items: [
              { value: 'equivalence', label: 'Equivalencia entre asignaturas' },
              { value: 'credits', label: 'Reconocimiento de créditos' },
            ],
          },
          {
            group: 'Pases',
            items: [
              { value: 'transfer', label: 'Pase de carrera' },
              { value: 'foreign', label: 'Pase desde una universidad extranjera', disabled: true },
            ],
          },
        ]}
      />
      <Select label="Carrera de origen" data={degrees} error="Elegí la carrera de origen" required />
      <Select label="Carrera actual" data={degrees} defaultValue="Ingeniería en Informática" disabled />
    </Stack>
  ),
  parameters: {
    docs: {
      description: { story: 'Opciones agrupadas, una deshabilitada, con error y deshabilitado.' },
    },
  },
}

export const MultiSelectStory: StoryObj = {
  name: 'MultiSelect',
  render: () => (
    <MultiSelect
      label="Asignaturas a reconocer"
      description="Elegí todas las que querés incluir"
      placeholder="Buscá una asignatura"
      data={subjects}
      defaultValue={['Análisis Matemático II', 'Física I']}
      searchable
      clearable
      nothingFoundMessage="No hay asignaturas que coincidan"
    />
  ),
  parameters: {
    docs: {
      description: { story: 'Varias opciones de una lista larga, como las asignaturas de una equivalencia.' },
    },
  },
}

export const AutocompleteStory: StoryObj = {
  name: 'Autocomplete',
  render: () => (
    <Autocomplete
      label="Universidad de origen"
      description="Si no está en la lista, escribila completa"
      placeholder="Universidad Nacional de La Plata"
      data={[
        'Universidad Nacional de La Plata',
        'Universidad Nacional de Córdoba',
        'Universidad Nacional del Sur',
        'Universidad Tecnológica Nacional',
        'Universidad de San Andrés',
      ]}
    />
  ),
  parameters: {
    docs: {
      description: {
        story: 'Texto libre con sugerencias: el valor puede no estar en la lista, como una universidad de origen.',
      },
    },
  },
}

export const TagsInputStory: StoryObj = {
  name: 'TagsInput',
  render: () => (
    <TagsInput
      label="Códigos de asignatura"
      description="Escribí un código y presioná Enter"
      placeholder="61.03"
      defaultValue={['61.03', '62.01']}
      data={['61.03', '61.08', '62.01', '75.41']}
      clearable
    />
  ),
  parameters: {
    docs: {
      description: { story: 'Varios valores libres, con sugerencias opcionales, como códigos de asignatura.' },
    },
  },
}

export const NativeSelectStory: StoryObj = {
  name: 'NativeSelect',
  render: () => (
    <Stack>
      <NativeSelect label="Cuatrimestre" data={['Primer cuatrimestre', 'Segundo cuatrimestre', 'Curso de verano']} />
      <NativeSelect
        label="Año de ingreso"
        data={['2023', '2024', '2025', '2026']}
        error="Elegí el año de ingreso a la carrera"
      />
    </Stack>
  ),
  parameters: {
    docs: {
      description: {
        story:
          'Una lista corta con el selector del sistema, que en el celular es más cómodo que un desplegable propio.',
      },
    },
  },
}

const institutions: CascaderOption[] = [
  {
    value: 'uba',
    label: 'Universidad de Buenos Aires',
    children: [
      {
        value: 'fiuba',
        label: 'Facultad de Ingeniería',
        children: [
          { value: 'fiuba-informatica', label: 'Ingeniería en Informática' },
          { value: 'fiuba-civil', label: 'Ingeniería Civil' },
          { value: 'fiuba-electronica', label: 'Ingeniería Electrónica' },
        ],
      },
      {
        value: 'fcen',
        label: 'Facultad de Ciencias Exactas y Naturales',
        children: [
          { value: 'fcen-computacion', label: 'Licenciatura en Ciencias de la Computación' },
          { value: 'fcen-fisica', label: 'Licenciatura en Ciencias Físicas' },
        ],
      },
    ],
  },
  {
    value: 'utn',
    label: 'Universidad Tecnológica Nacional',
    children: [
      {
        value: 'frba',
        label: 'Facultad Regional Buenos Aires',
        children: [{ value: 'frba-sistemas', label: 'Ingeniería en Sistemas de Información' }],
      },
    ],
  },
]

export const CascaderStory: StoryObj = {
  name: 'Cascader',
  render: () => (
    <Cascader
      label="Carrera de origen"
      description="Elegí la universidad, la facultad y la carrera"
      placeholder="Elegí una carrera"
      data={institutions}
      separator=" / "
      searchable
      clearable
      nothingFoundMessage="No hay carreras que coincidan"
    />
  ),
  parameters: {
    docs: {
      description: {
        story: 'Una opción dentro de niveles anidados, como la carrera de origen de un pase dentro de su universidad.',
      },
    },
  },
}

const curriculum: TreeNodeData[] = [
  {
    value: 'basic',
    label: 'Ciclo básico',
    children: [
      { value: 'am2', label: 'Análisis Matemático II' },
      { value: 'alg2', label: 'Álgebra II' },
      { value: 'fis1', label: 'Física I' },
    ],
  },
  {
    value: 'professional',
    label: 'Ciclo profesional',
    children: [
      { value: 'algo', label: 'Algoritmos y Estructuras de Datos' },
      { value: 'db', label: 'Base de Datos' },
      { value: 'os', label: 'Sistemas Operativos' },
    ],
  },
]

export const TreeSelectStory: StoryObj = {
  name: 'TreeSelect',
  render: () => (
    <TreeSelect
      mode="checkbox"
      label="Asignaturas del plan de estudios"
      description="Marcar un ciclo marca todas sus asignaturas"
      placeholder="Elegí las asignaturas"
      data={curriculum}
      defaultValue={['am2']}
      defaultExpandAll
      searchable
      nothingFoundMessage="No hay asignaturas que coincidan"
    />
  ),
  parameters: {
    docs: {
      description: {
        story: 'Opciones en un árbol, como las asignaturas de un plan de estudios agrupadas por ciclo.',
      },
    },
  },
}

function PillsInputDemo() {
  const [values, setValues] = useState(['Análisis Matemático II', 'Física I'])
  const [search, setSearch] = useState('')

  const add = () => {
    const value = search.trim()
    if (value && !values.includes(value)) {
      setValues((current) => [...current, value])
    }
    setSearch('')
  }

  return (
    <PillsInput label="Asignaturas aprobadas" description="Escribí el nombre y presioná Enter">
      <Pill.Group>
        {values.map((value) => (
          <Pill
            key={value}
            withRemoveButton
            removeButtonProps={{ 'aria-label': `Quitar ${value}` }}
            onRemove={() => setValues((current) => current.filter((item) => item !== value))}
          >
            {value}
          </Pill>
        ))}
        <PillsInput.Field
          value={search}
          placeholder="Agregar asignatura"
          onChange={(event) => setSearch(event.currentTarget.value)}
          onKeyDown={(event) => {
            if (event.key === 'Enter') {
              event.preventDefault()
              add()
            }
          }}
        />
      </Pill.Group>
    </PillsInput>
  )
}

export const PillsInputStory: StoryObj = {
  name: 'PillsInput',
  render: () => <PillsInputDemo />,
  parameters: {
    docs: {
      description: {
        story:
          'La base de `MultiSelect` y `TagsInput`, con `Pill` para cada valor. Para un campo propio de varios valores, cuando ninguno de los dos alcanza.',
      },
    },
  },
}

function SearchableSubjectSelect() {
  const [value, setValue] = useState<string | null>(null)
  const [search, setSearch] = useState('')

  const combobox = useCombobox({
    onDropdownClose: () => combobox.resetSelectedOption(),
    onDropdownOpen: () => combobox.selectFirstOption(),
  })

  // With the selected label in the input, show the whole list instead of filtering by it.
  const shouldFilter = subjects.every((item) => item !== search)
  const filtered = shouldFilter
    ? subjects.filter((item) => item.toLowerCase().includes(search.toLowerCase().trim()))
    : subjects

  return (
    <Combobox
      store={combobox}
      onOptionSubmit={(selected) => {
        setValue(selected)
        setSearch(selected)
        combobox.closeDropdown()
      }}
    >
      <Combobox.Target>
        <InputBase
          label="Asignatura de la FIUBA"
          description="La asignatura que querés que se reconozca"
          placeholder="Buscá una asignatura"
          rightSection={<Combobox.Chevron />}
          rightSectionPointerEvents="none"
          value={search}
          onChange={(event) => {
            combobox.openDropdown()
            combobox.updateSelectedOptionIndex()
            setSearch(event.currentTarget.value)
          }}
          onClick={() => combobox.openDropdown()}
          onFocus={() => combobox.openDropdown()}
          onBlur={() => {
            combobox.closeDropdown()
            setSearch(value ?? '')
          }}
        />
      </Combobox.Target>
      <Combobox.Dropdown>
        <Combobox.Options>
          {filtered.length > 0 ? (
            filtered.map((item) => (
              <Combobox.Option value={item} key={item} active={item === value}>
                {item}
              </Combobox.Option>
            ))
          ) : (
            <Combobox.Empty>No hay asignaturas que coincidan</Combobox.Empty>
          )}
        </Combobox.Options>
      </Combobox.Dropdown>
    </Combobox>
  )
}

export const ComboboxStory: StoryObj = {
  name: 'Combobox',
  render: () => <SearchableSubjectSelect />,
  parameters: {
    docs: {
      description: {
        story:
          'Las piezas con las que se arman `Select` y los demás. Se usa solo para un selector que ninguno resuelve; este ejemplo es un select con búsqueda.',
      },
    },
  },
}
