import { Box, Paper } from '@mantine/core'
import { Children, type ReactNode } from 'react'

// ProcedureRows holds the rows in a single container, with a divider between
// them.
export function ProcedureRows({ label, children }: { label: string; children: ReactNode[] }) {
  return (
    <Paper withBorder variant="rows" component="ul" aria-label={label} p={0} m={0}>
      {Children.map(children, (row) => (
        <Box component="li">{row}</Box>
      ))}
    </Paper>
  )
}
