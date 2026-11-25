import type { Meta, StoryObj } from '@storybook/react-vite'
import { useState } from 'react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { Figure } from './Figure'
import { Figures } from '../Figures/Figures'

const meta: Meta = {
  title: 'SIGTA/Figure',
  decorators: [(Story) => <RouterProvider router={createMemoryRouter([{ path: '*', element: <Story /> }])} />],
  parameters: {
    docs: {
      description: {
        component:
          'Un número grande que cuenta hasta su valor y lo que cuenta, en una celda de una fila de cifras. Toda la celda es un control: un enlace a donde se cuenta la cifra, o un botón que filtra por ella. Con el puntero sube un tinte desde su pie, el número se eleva y la flecha avanza. Elegida, lleva la marca de la selección, que viaja desde la cifra anterior; con movimiento reducido cambia sin desplazarse.',
      },
    },
  },
}

export default meta

export const Link: StoryObj = {
  name: 'Enlace',
  render: () => (
    <Figures label="Resumen">
      <Figure value={4} label="Requieren acción" to="/inbox" />
      <Figure value={11} label="En curso" to="/inbox?view=all" />
    </Figures>
  ),
}

function Filters() {
  const [selected, setSelected] = useState('mine')
  const options = [
    { key: 'mine', label: 'Acción requerida', value: 4 },
    { key: 'office', label: 'En un área', value: 6 },
  ]
  return (
    <Figures label="Resumen por estado">
      {options.map((option) => (
        <Figure
          key={option.key}
          value={option.value}
          label={option.label}
          selected={selected === option.key}
          onSelect={() => setSelected(option.key)}
        />
      ))}
    </Figures>
  )
}

export const Filter: StoryObj = {
  name: 'Filtro',
  render: () => <Filters />,
}
