import { Button, Stack, Text } from '@mantine/core'
import type { Meta, StoryObj } from '@storybook/react-vite'
import { useState } from 'react'
import { CountUp } from './CountUp'

const meta: Meta<typeof CountUp> = {
  title: 'SIGTA/CountUp',
  component: CountUp,
  parameters: {
    docs: {
      description: {
        component:
          'Una cifra que cuenta hasta su valor al aparecer y del valor anterior al nuevo cuando cambia, en `motion.count`. Un lector de pantalla lee solo el valor. Con movimiento reducido muestra el valor directamente.',
      },
    },
  },
}

export default meta

export const Figure: StoryObj<typeof CountUp> = {
  name: 'CountUp',
  args: { value: 24 },
  render: (args) => (
    <Text fz="xl" fw={600}>
      <CountUp {...args} />
    </Text>
  ),
}

function Changing() {
  const [value, setValue] = useState(3)
  return (
    <Stack align="flex-start" gap="sm">
      <Text fz="xl" fw={600}>
        <CountUp value={value} />
      </Text>
      <Button variant="default" onClick={() => setValue((current) => (current === 3 ? 17 : 3))}>
        Cambiar la cifra
      </Button>
    </Stack>
  )
}

export const Change: StoryObj<typeof CountUp> = {
  name: 'CountUp, al cambiar',
  render: () => <Changing />,
}
