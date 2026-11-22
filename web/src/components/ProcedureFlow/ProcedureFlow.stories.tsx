import type { Meta, StoryObj } from '@storybook/react-vite'
import { BlueprintPanel } from '../BlueprintPanel/BlueprintPanel'
import { ProcedureFlow } from './ProcedureFlow'

const meta: Meta<typeof ProcedureFlow> = {
  title: 'SIGTA/ProcedureFlow',
  component: ProcedureFlow,
  parameters: {
    docs: {
      description: {
        component:
          'El recorrido de un trámite derivado, en la portada: se traza una vez en `motion.draw` y después un punto lo recorre en `motion.flow`, de la solicitud a la resolución, con una pausa de `motion.rest`; cada parada se ilumina cuando el punto llega. Va sobre el plano (`BlueprintPanel`) y los lectores de pantalla lo saltean. Con movimiento reducido queda trazado y quieto.',
      },
    },
  },
}

export default meta

export const OnBlueprint: StoryObj<typeof ProcedureFlow> = {
  name: 'ProcedureFlow, sobre el plano',
  render: () => (
    <BlueprintPanel radius="md" p="xl" maw="var(--sigta-measure)">
      <ProcedureFlow />
    </BlueprintPanel>
  ),
}
