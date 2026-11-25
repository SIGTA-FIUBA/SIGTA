import { Stack, Title } from '@mantine/core'
import type { ReactNode } from 'react'

// Part is a part of a page: a rule above, its title and what
// it holds, with no box around it.
export function Part({ title, children }: { title: string; children: ReactNode }) {
  return (
    <Stack component="section" variant="part" gap="md" aria-label={title}>
      <Title order={2}>{title}</Title>
      {children}
    </Stack>
  )
}
