import { ActionBar, Button, Checkbox, Stack, Text, Toolbar } from '@mantine/core'
import { type ReactNode, useState } from 'react'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { Icon } from '../../components/Icon/Icon'

// Bars that group the actions on a list. ActionBar appears while there is a
// selection; Toolbar stays above the list.

const meta: Meta = {
  title: 'Componentes/Barras de acciones',
  parameters: {
    docs: {
      description: {
        component:
          'Agrupan las acciones que se hacen sobre una lista de trámites. Los botones que solo tienen un ícono llevan `aria-label`.',
      },
    },
  },
}

export default meta

const procedures = [
  { id: 'EQ-2026-00042', label: 'Equivalencia entre asignaturas, Martina Quiroga Leal' },
  { id: 'RC-2026-00017', label: 'Reconocimiento de créditos, Tomás Echeverría Paz' },
  { id: 'PC-2026-00008', label: 'Pase de carrera, Julieta Arrieta Sosa' },
]

function ActionBarDemo() {
  const [selected, setSelected] = useState<string[]>(['EQ-2026-00042'])
  return (
    <>
      <Checkbox.Group label="Trámites de la bandeja" value={selected} onChange={setSelected}>
        <Stack gap="xs" mt="xs">
          {procedures.map((procedure) => (
            <Checkbox key={procedure.id} value={procedure.id} label={`${procedure.id}, ${procedure.label}`} />
          ))}
        </Stack>
      </Checkbox.Group>
      <ActionBar
        opened={selected.length > 0}
        onClose={() => setSelected([])}
        closeOnEscape
        aria-label="Acciones sobre los trámites seleccionados"
      >
        <Text size="sm">
          {selected.length === 1 ? '1 trámite seleccionado' : `${selected.length} trámites seleccionados`}
        </Text>
        <ActionBar.Divider />
        <Button variant="default" size="xs">
          Asignar
        </Button>
        <Button variant="default" size="xs">
          Descargar documentación
        </Button>
        <ActionBar.CloseButton aria-label="Quitar la selección" />
      </ActionBar>
    </>
  )
}

export const ActionBarStory: StoryObj = {
  name: 'ActionBar',
  render: () => <ActionBarDemo />,
  parameters: {
    docs: {
      description: {
        story:
          'Aparece abajo de la pantalla mientras hay trámites seleccionados en la bandeja, con las acciones que se aplican a todos. Se cierra con la cruz o con Escape.',
      },
    },
  },
}

// A filter that turns on and off. An icon-only one takes its label as aria-label.
function FilterToggle({ label, children }: { label?: string; children: ReactNode }) {
  const [active, setActive] = useState(false)
  return (
    <Toolbar.Toggle
      autoWidth={label === undefined}
      aria-label={label}
      active={active}
      onClick={() => setActive((value) => !value)}
    >
      {children}
    </Toolbar.Toggle>
  )
}

export const ToolbarStory: StoryObj<typeof Toolbar> = {
  name: 'Toolbar',
  render: (args) => <Toolbar {...args} />,
  args: {
    'aria-label': 'Filtros de la bandeja',
    size: 'sm',
    variant: 'light',
    withBorder: true,
    orientation: 'horizontal',
    children: (
      <>
        <Toolbar.ToggleGroup type="single" defaultValue="all">
          <Toolbar.ToggleItem value="all" autoWidth>
            Todos
          </Toolbar.ToggleItem>
          <Toolbar.ToggleItem value="in_office" autoWidth>
            En un área
          </Toolbar.ToggleItem>
          <Toolbar.ToggleItem value="referred" autoWidth>
            Derivado
          </Toolbar.ToggleItem>
        </Toolbar.ToggleGroup>
        <Toolbar.Divider />
        <Toolbar.Group>
          <FilterToggle>Con acción requerida</FilterToggle>
          <FilterToggle label="Solo los trámites asignados">
            <Icon name="inbox" />
          </FilterToggle>
        </Toolbar.Group>
      </>
    ),
  },
  argTypes: {
    size: { control: 'select', options: ['xs', 'sm', 'md', 'lg', 'xl'] },
    variant: { control: 'select', options: ['light', 'filled'] },
    orientation: { control: 'select', options: ['horizontal', 'vertical'] },
    children: { control: false },
  },
  parameters: {
    docs: {
      description: {
        story:
          'Sobre la bandeja, para filtrar la lista por estado. `ToggleGroup` elige una opción entre varias, `Toggle` prende y apaga un filtro y `Divider` separa los grupos.',
      },
    },
  },
}
