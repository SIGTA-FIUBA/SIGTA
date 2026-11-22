import { Paper } from '@mantine/core'
import { Children, type ReactNode } from 'react'
import { Reveal } from '../Reveal/Reveal'

// ProcedureRows holds the rows in a single container, with a divider between
// them. The rows arrive one after the other, from the top, so the eye reads
// the list in its order.
export function ProcedureRows({ label, children }: { label: string; children: ReactNode[] }) {
  return (
    <Paper withBorder variant="rows" component="ul" aria-label={label} p={0} m={0}>
      {Children.map(children, (row, index) => (
        <Reveal as="li" index={index}>
          {row}
        </Reveal>
      ))}
    </Paper>
  )
}
