import { Stack, Text } from '@mantine/core'
import type { ReactNode } from 'react'

type EmptyStateProps = {
  title: string
  text: string
  // action is the primary button, when the page has no other.
  action?: ReactNode
}

// EmptyState takes the place of a list with nothing to show: a title, one
// line and what to do next. No illustration and no box.
export function EmptyState({ title, text, action }: EmptyStateProps) {
  return (
    <Stack gap="xxs" align="flex-start">
      <Text fw={600}>{title}</Text>
      <Text size="sm" c="dimmed" maw="var(--sigta-measure)">
        {text}
      </Text>
      {action && <Stack mt="sm">{action}</Stack>}
    </Stack>
  )
}
