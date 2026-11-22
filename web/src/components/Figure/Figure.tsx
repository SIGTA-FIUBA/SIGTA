import { Box, Group, Paper, Stack, Text } from '@mantine/core'
import { motion } from 'motion/react'
import { Link } from 'react-router'
import { move } from '../../theme/motion'
import { CountUp } from '../CountUp/CountUp'
import { Icon } from '../Icon/Icon'

type FigureProps = {
  value: number
  label: string
} & (
  | { to: string; onSelect?: never; selected?: never }
  // A figure that filters what is below it: pressed, it carries the mark of
  // the selection, which travels from the figure left to the one chosen.
  | { to?: never; onSelect: () => void; selected: boolean }
)

// Figure is a large number that counts up to its value and what it counts, in
// a cell of a row of figures. The whole cell is a control: a link to where the
// figure is counted, or a button that filters by it. Under the pointer a tint
// rises from its foot, the number lifts and the arrow moves forward.
export function Figure({ value, label, to, onSelect, selected }: FigureProps) {
  const content = (
    <Group justify="space-between" align="flex-end" wrap="nowrap" gap="sm" pos="relative">
      <Stack gap="xs" miw={0}>
        <Text component="div" variant="figure" c="fiuba.9" data-figure>
          <CountUp value={value} />
        </Text>
        <Text size="sm" c="gray.7">
          {label}
        </Text>
      </Stack>
      {to && (
        <Box data-chevron aria-hidden>
          <Icon name="chevron-right" size={16} />
        </Box>
      )}
    </Group>
  )
  if (to) {
    return (
      <Paper component={Link} to={to} variant="figure" radius={0} px={{ base: 'md', md: 'lg' }} py="lg">
        {content}
      </Paper>
    )
  }
  return (
    <Paper
      component="button"
      type="button"
      onClick={onSelect}
      aria-pressed={selected}
      variant="figure"
      radius={0}
      px={{ base: 'md', md: 'lg' }}
      py="lg"
    >
      {selected && (
        <Paper
          variant="selection"
          radius={0}
          renderRoot={(props) => <motion.div {...props} layoutId="figure-selection" transition={move} />}
        />
      )}
      {content}
    </Paper>
  )
}
