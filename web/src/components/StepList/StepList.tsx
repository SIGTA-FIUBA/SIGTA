import { List, Text, Title } from '@mantine/core'

export type Step = { title: string; text: string }

// StepList shows how something is done, in order: an ordered list in three
// columns, without boxes, each step a heading under the title of its
// section. The number is in the action color, without a circle.
export function StepList({ steps }: { steps: Step[] }) {
  return (
    <List type="ordered" variant="steps">
      {steps.map((step, index) => (
        <List.Item key={step.title}>
          <Text fw={600} c="fiuba.7">
            {index + 1}
          </Text>
          <Title order={3} size="h4" mt="xxs">
            {step.title}
          </Title>
          <Text size="sm" c="dimmed" mt="xxs">
            {step.text}
          </Text>
        </List.Item>
      ))}
    </List>
  )
}
