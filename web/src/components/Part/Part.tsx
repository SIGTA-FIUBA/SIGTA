import { Stack, Text, Title } from '@mantine/core'
import type { ReactNode } from 'react'

// Part is a part of a page: a rule above, its title and what
// it holds, with no box around it.
export function Part({ title, lead, children }: { title: string; lead?: string; children: ReactNode }) {
  return (
    <Stack component="section" variant="part" gap="md" aria-label={title}>
      <Stack gap="xxs">
        <Title order={2}>{title}</Title>
        {lead && (
          <Text size="sm" c="gray.7">
            {lead}
          </Text>
        )}
      </Stack>
      {children}
    </Stack>
  )
}
