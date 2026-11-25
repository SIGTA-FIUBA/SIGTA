import { Grid, Text } from '@mantine/core'
import type { ReactNode } from 'react'

// Fact is one fact of Facts: its label as a dt and its value as a dd.
export function Fact({ label, split, children }: { label: string; split?: boolean; children: ReactNode }) {
  return (
    <Grid component="div" gap={split ? { base: 'xxs', sm: 'md' } : 'xxs'} py="sm">
      <Grid.Col span={split ? { base: 12, sm: 4 } : 12}>
        <Text component="dt" size="sm" c="gray.7">
          {label}
        </Text>
      </Grid.Col>
      <Grid.Col component="dd" span={split ? { base: 12, sm: 8 } : 12} m={0}>
        {children}
      </Grid.Col>
    </Grid>
  )
}
