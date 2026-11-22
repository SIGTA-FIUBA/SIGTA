import { Box, Flex, Group, Paper, Stack, Text } from '@mantine/core'
import { AnimatePresence, motion } from 'motion/react'
import { Children, useId, useState, type ReactNode } from 'react'
import { move } from '../../theme/motion'
import { Icon } from '../Icon/Icon'
import { rowColumns } from '../ProcedureRow/columns'
import { Reveal } from '../Reveal/Reveal'

// ProcedureRows sets the rows as a table in a book, with no box: a rule above,
// one between the rows and one below, and on a wide screen the headings of
// its four columns. One gray mark glides under the row that the pointer or
// the keyboard is on, so the list answers as a whole. The rows arrive one
// after the other, from the top, so the eye reads the list in its order.
export function ProcedureRows({
  label,
  columns,
  children,
}: {
  label: string
  // columns are the headings of the four columns of the rows.
  columns?: readonly [string, string, string, string]
  children: ReactNode[]
}) {
  const mark = useId()
  const [current, setCurrent] = useState<number | null>(null)
  return (
    <Stack gap="xs">
      {columns && (
        // The headings repeat for the eye what each row says to a screen
        // reader, so they stay out of the accessibility tree.
        <Group wrap="nowrap" gap="md" visibleFrom="md" aria-hidden>
          <Flex gap="lg" flex={1} miw={0}>
            {columns.map((column, index) => (
              <Text key={column} variant="column" flex={rowColumns[index]} ta={index === 3 ? 'right' : undefined}>
                {column}
              </Text>
            ))}
          </Flex>
          {/* The place of the arrow of the rows. */}
          <Box opacity={0} display="flex">
            <Icon name="chevron-right" size={16} />
          </Box>
        </Group>
      )}
      <Paper
        variant="rows"
        component="ul"
        aria-label={label}
        radius={0}
        p={0}
        m={0}
        onPointerLeave={() => setCurrent(null)}
        onBlur={() => setCurrent(null)}
      >
        {Children.map(children, (row, index) => (
          <Reveal as="li" index={index} onPointerEnter={() => setCurrent(index)} onFocus={() => setCurrent(index)}>
            <AnimatePresence>
              {current === index && (
                <Paper
                  variant="hover"
                  aria-hidden
                  renderRoot={(props) => (
                    <motion.div
                      {...props}
                      layoutId={mark}
                      transition={move}
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                    />
                  )}
                />
              )}
            </AnimatePresence>
            {row}
          </Reveal>
        ))}
      </Paper>
    </Stack>
  )
}
