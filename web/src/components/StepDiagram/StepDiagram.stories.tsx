import type { Meta, StoryObj } from '@storybook/react-vite'
import { Text } from '@mantine/core'
import { useState } from 'react'
import { StepDiagram, type DiagramStep } from './StepDiagram'

const meta: Meta = {
  title: 'SIGTA/StepDiagram',
  parameters: {
    docs: {
      description: {
        component:
          'El recorrido de un tipo de trámite como máquina de estados, sobre un lienzo punteado: cada paso es una caja y cada salida, una línea. Se destacan las líneas y la caja del paso elegido; las devoluciones al titular van en naranja.',
      },
    },
  },
}

export default meta

const steps: DiagramStep[] = [
  {
    id: 'request',
    next: 'review',
    name: 'Solicitud',
    kicker: 'Titular',
    detail: 'Completa el formulario',
    applicant: true,
    problem: false,
  },
  {
    id: 'review',
    outcomes: [
      { name: 'Aprobado', next: 'resolution', send_back: false },
      { name: 'Falta documentación', next: 'request', send_back: true },
    ],
    name: 'Revisión',
    kicker: 'Dirección de Alumnos',
    detail: 'Decide si sigue',
    applicant: false,
    problem: false,
  },
  {
    id: 'resolution',
    name: 'Resolución',
    kicker: 'Falta el área',
    detail: 'Adjunta la resolución',
    applicant: false,
    problem: true,
  },
]

function Diagram() {
  const [selected, setSelected] = useState('review')
  return (
    <StepDiagram
      steps={steps}
      selected={selected}
      onSelect={setSelected}
      panel={<Text>Paso elegido: {steps.find((step) => step.id === selected)?.name}</Text>}
    />
  )
}

export const Route: StoryObj = {
  name: 'Recorrido',
  render: () => <Diagram />,
}
