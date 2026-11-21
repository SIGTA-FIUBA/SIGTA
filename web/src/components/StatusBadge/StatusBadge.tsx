import { Badge } from '@mantine/core'
import type { ReactNode } from 'react'

// The tone says who has the procedure; the text says it, so the color is
// never the only sign. The names of the states are the Secretaría's.
export type StatusTone = 'office' | 'external' | 'applicant' | 'closed' | 'done' | 'error'

const colors: Record<StatusTone, string> = {
  office: 'fiuba',
  external: 'orange',
  applicant: 'gray',
  closed: 'gray',
  done: 'green',
  error: 'red',
}

// StatusBadge shows the state of a procedure: caption in bold, the 700 shade
// on the 50 tint of its color.
export function StatusBadge({ tone, children }: { tone: StatusTone; children: ReactNode }) {
  return <Badge color={colors[tone]}>{children}</Badge>
}
