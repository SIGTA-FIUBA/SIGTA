import { Stack } from '@mantine/core'
import type { ReactNode } from 'react'

// Facts is a list of what something is, as a definition list between
// rules, with no box: one fact under the other with a hairline between them.
// Split, each label takes the first third of the line and its value the rest;
// stacked, the label goes above, for a narrow column.
export function Facts({ label, children }: { label: string; children: ReactNode }) {
  return (
    <Stack component="dl" variant="facts" gap={0} m={0} aria-label={label}>
      {children}
    </Stack>
  )
}
