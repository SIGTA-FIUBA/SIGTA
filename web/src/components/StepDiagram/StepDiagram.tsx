import { Badge, Box, Grid, Group, Paper, ScrollArea, Text, UnstyledButton } from '@mantine/core'
import { useElementSize } from '@mantine/hooks'
import { useEffect, useRef, type ReactNode } from 'react'
import { Placed } from './Placed'
import { stepLayout, type LayoutStep } from './stepLayout'

// A step as the diagram shows it: who does it (or that it has a problem)
// above its name, and what it does below.
export type DiagramStep = LayoutStep & {
  name: string
  kicker: string
  detail: string
  applicant: boolean
  problem: boolean
}

// The lines are drawn in SVG with the colors of the theme, through Mantine's
// CSS variables.
const lineColor = {
  plain: 'var(--mantine-color-gray-5)',
  selected: 'var(--mantine-primary-color-filled)',
  sendBack: 'var(--mantine-color-orange-8)',
}

// A diagram narrower than this goes compact: it would not fit its labels.
const compactBelow = 560

// StepDiagram draws the route of a procedure type as a state machine, on a
// dotted canvas: each step is a box, from the start to the close, and each
// way out of a step is a line. The lines and box of the selected step stand
// out. panel goes beside the diagram, floating on the canvas, or under it on
// a narrow screen, and toolbar goes under the diagram.
export function StepDiagram({
  steps,
  selected,
  onSelect,
  panel,
  toolbar,
}: {
  steps: DiagramStep[]
  selected: string
  onSelect: (stepId: string) => void
  panel: ReactNode
  toolbar?: ReactNode
}) {
  const { ref: area, width } = useElementSize()
  const layout = stepLayout(steps, { compact: width > 0 && width < compactBelow })
  // A diagram wider than the screen opens on its center line, where the
  // steps are, and leaves the sides to scroll.
  const viewport = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const box = viewport.current
    if (box) {
      box.scrollLeft = (box.scrollWidth - box.clientWidth) / 2
    }
  }, [layout.width])
  const colorOf = (from: string, sendBack: boolean) =>
    sendBack ? lineColor.sendBack : from === selected ? lineColor.selected : lineColor.plain

  return (
    <Paper
      variant="canvas"
      withBorder
      shadow="xs"
      component="section"
      aria-label="Recorrido del tipo de trámite"
      p="md"
    >
      <Grid gap="md">
        {/* miw keeps the column at its share: it would grow to the width of the diagram. */}
        <Grid.Col span={{ base: 12, md: 8 }} miw={0}>
          <Box ref={area}>
            <ScrollArea type="auto" viewportRef={viewport} offsetScrollbars>
              <Placed frame width={layout.width} height={layout.height}>
                <Placed>
                  <svg width={layout.width} height={layout.height} aria-hidden="true">
                    {layout.edges.map((edge, index) => {
                      const color = colorOf(edge.from, edge.sendBack)
                      return (
                        <g key={index}>
                          <path
                            d={edge.path}
                            fill="none"
                            stroke={color}
                            strokeWidth={edge.from === selected || edge.sendBack ? 2.5 : 2}
                            strokeDasharray={edge.kind === 'back' ? '6 5' : undefined}
                          />
                          <path d={edge.head} fill={color} />
                        </g>
                      )
                    })}
                  </svg>
                </Placed>
                {layout.edges.map(
                  (edge, index) =>
                    edge.label && (
                      <Placed
                        key={index}
                        x={edge.labelX}
                        y={edge.labelY}
                        anchor={edge.labelAnchor}
                        maxWidth={layout.labelWidth}
                      >
                        <Badge
                          variant={edge.sendBack ? 'light' : edge.from === selected ? 'filled' : 'default'}
                          color={edge.sendBack ? 'orange' : edge.from === selected ? 'fiuba' : 'dark'}
                          tt="none"
                          maw="100%"
                          title={edge.label}
                        >
                          {edge.sendBack ? `↩ ${edge.label}` : edge.label}
                        </Badge>
                      </Placed>
                    ),
                )}
                {steps.length > 0 && (
                  <Placed x={layout.centerX} y={14} anchor="middle">
                    <Badge color="dark" tt="none">
                      Inicio
                    </Badge>
                  </Placed>
                )}
                <Placed x={layout.centerX} y={layout.closeTop} anchor="middle">
                  <Badge variant="outline" color="dark" tt="none">
                    Cierre
                  </Badge>
                </Placed>
                {steps.map((step, index) => {
                  const isSelected = step.id === selected
                  return (
                    <Placed
                      key={step.id}
                      x={layout.centerX}
                      y={layout.nodeTop(index)}
                      width={layout.nodeWidth}
                      height={layout.nodeHeight}
                      anchor="middle"
                    >
                      <UnstyledButton
                        w="100%"
                        h="100%"
                        aria-pressed={isSelected}
                        aria-label={`Paso ${index + 1}: ${step.name}. ${step.kicker}. ${step.detail}`}
                        onClick={() => onSelect(step.id)}
                      >
                        <Paper
                          variant="step"
                          radius="md"
                          bg={isSelected ? 'fiuba.0' : 'white'}
                          mod={{ selected: isSelected, problem: step.problem }}
                        >
                          <Group
                            justify="space-between"
                            px="sm"
                            bg={step.problem ? 'red.0' : step.applicant ? 'fiuba.0' : 'gray.1'}
                          >
                            <Text
                              component="span"
                              size="xs"
                              fw={700}
                              tt="uppercase"
                              lh="xl"
                              c={step.problem ? 'red.8' : step.applicant ? 'fiuba.8' : 'gray.8'}
                            >
                              {step.kicker}
                            </Text>
                            <Text component="span" size="xs" fw={700} lh="xl" c="dimmed">
                              {index + 1}
                            </Text>
                          </Group>
                          <Box px="sm" pt="xs">
                            <Text fw={600} lh="xs" lineClamp={2}>
                              {step.name}
                            </Text>
                            <Text size="xs" c="dimmed" truncate="end">
                              {step.detail}
                            </Text>
                          </Box>
                        </Paper>
                      </UnstyledButton>
                    </Placed>
                  )
                })}
              </Placed>
            </ScrollArea>
          </Box>
          {toolbar && (
            <Group justify="center" mt="sm">
              {toolbar}
            </Group>
          )}
        </Grid.Col>
        <Grid.Col span={{ base: 12, md: 4 }}>
          <Paper shadow="md" radius="md" p="lg" withBorder>
            {panel}
          </Paper>
        </Grid.Col>
      </Grid>
    </Paper>
  )
}
