import { Group, Stack, Text, Title } from '@mantine/core'
import type { ReactNode } from 'react'

type PageHeaderProps = {
  title: string
  // description says what the page is for, in one or two lines.
  description?: string
  // action is the primary action of the page, the only primary button.
  action?: ReactNode
}

// PageHeader opens every page: its title, an optional description and, on the
// right, the page's primary action. Nothing goes above the title: the logo
// already says where the person is.
export function PageHeader({ title, description, action }: PageHeaderProps) {
  return (
    <Group justify="space-between" align="flex-start" gap="md">
      <Stack gap="xs">
        <Title order={1}>{title}</Title>
        {description && (
          <Text c="dimmed" maw="var(--sigta-measure)">
            {description}
          </Text>
        )}
      </Stack>
      {action}
    </Group>
  )
}
