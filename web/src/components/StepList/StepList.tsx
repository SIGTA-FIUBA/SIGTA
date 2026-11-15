import { List, Text } from '@mantine/core'

export type Step = { title: string; text: string }

// StepList shows how something is done, in order: an ordered list in three
// columns, without boxes. The number is in the action color, without a circle.
export function StepList({ steps }: { steps: Step[] }) {
  return (
    <List type="ordered" variant="steps">
      {steps.map((step, index) => (
        <List.Item key={step.title}>
          <Text fw={700} c="fiuba.7">
            {index + 1}
          </Text>
          <Text fw={700} mt="xxs">
            {step.title}
          </Text>
          <Text size="sm" c="dimmed" mt="xxs">
            {step.text}
          </Text>
        </List.Item>
      ))}
    </List>
  )
}
