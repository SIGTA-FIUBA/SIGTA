import { SimpleGrid } from '@mantine/core'
import type { ReactNode } from 'react'

// Figures is a row of figures as wide as its column, between two rules, with
// a hairline between the cells. Its columns divide the figures, so the row
// always closes: two figures in two columns, four in two on a phone and in
// four from md on.
export function Figures({ label, children }: { label: string; children: ReactNode[] }) {
  const four = children.length === 4
  return (
    <SimpleGrid variant="figures" cols={four ? { base: 2, md: 4 } : children.length} spacing={0} aria-label={label}>
      {children}
    </SimpleGrid>
  )
}
