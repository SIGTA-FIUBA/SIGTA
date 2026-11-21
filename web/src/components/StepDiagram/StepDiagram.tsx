import { Box, Group, Paper, Text } from '@mantine/core'
import type { ReactNode } from 'react'
import { stepLayout, type LayoutStep } from './stepLayout'

// A step as the diagram shows it: who does it (or the problem it has) above
// its name, and what it does below.
export type DiagramStep = LayoutStep & {
  name: string
  kicker: string
  detail: string
  applicant: boolean
  problem: boolean
}

// The diagram is the one place of the interface drawn with coordinates, so
// it is the one place that styles elements inline (DESIGN.md). Every color
// still comes from the theme, through Mantine's CSS variables.
const color = {
  line: 'var(--mantine-color-gray-5)',
  selected: 'var(--mantine-primary-color-filled)',
  sendBack: 'var(--mantine-color-orange-8)',
  problem: 'var(--mantine-color-red-7)',
  dots: 'var(--mantine-color-gray-3)',
}

// StepDiagram draws the route of a procedure type as a state machine, on a
// dotted canvas: each step is a box, from the start to the close, and each
// way out of a step is a line. The lines and boxes of the selected step stand
// out. panel goes beside the diagram, floating on the canvas, and toolbar
// floats under it.
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
  const layout = stepLayout(steps)
  return (
    <Paper
      withBorder
      shadow="xs"
      aria-label="Recorrido del tipo de trámite"
      component="section"
      style={{
        backgroundColor: 'var(--mantine-color-gray-0)',
        backgroundImage: `radial-gradient(${color.dots} 1px, transparent 1px)`,
        backgroundSize: '20px 20px',
        overflow: 'hidden',
      }}
    >
      <Group align="flex-start" gap="md" p="md" wrap="wrap">
        <Box style={{ flex: '1 1 640px', minWidth: 0, overflowX: 'auto' }}>
          <div style={{ position: 'relative', width: layout.width, height: layout.height, margin: '0 auto' }}>
            <svg
              width={layout.width}
              height={layout.height}
              aria-hidden="true"
              style={{ position: 'absolute', inset: 0 }}
            >
              {layout.edges.map((edge, index) => {
                const stroke = edge.sendBack ? color.sendBack : edge.from === selected ? color.selected : color.line
                return (
                  <g key={index}>
                    <path
                      d={edge.path}
                      fill="none"
                      stroke={stroke}
                      strokeWidth={edge.from === selected || edge.sendBack ? 2.5 : 2}
                      strokeDasharray={edge.kind === 'back' ? '6 5' : undefined}
                    />
                    <path d={edge.head} fill={stroke} />
                  </g>
                )
              })}
            </svg>
            {layout.edges.map(
              (edge, index) =>
                edge.label && (
                  <Text
                    key={index}
                    component="span"
                    size="xs"
                    fw={600}
                    style={{
                      position: 'absolute',
                      left: edge.labelX,
                      top: edge.labelY,
                      transform: edge.labelAnchor === 'end' ? 'translateX(-100%)' : undefined,
                      whiteSpace: 'nowrap',
                      padding: '1px 10px',
                      borderRadius: 999,
                      background: 'var(--mantine-color-white)',
                      border: `1px solid ${edge.sendBack ? color.sendBack : edge.from === selected ? color.selected : color.line}`,
                      color: edge.sendBack ? color.sendBack : 'var(--mantine-color-black)',
                    }}
                  >
                    {edge.sendBack ? `↩ ${edge.label}` : edge.label}
                  </Text>
                ),
            )}
            {steps.length > 0 && (
              <Text component="span" size="xs" fw={600} c="white" style={terminal(layout.centerX, 14, true)}>
                Inicio
              </Text>
            )}
            <Text component="span" size="xs" fw={600} style={terminal(layout.centerX, layout.closeTop, false)}>
              Cierre
            </Text>
            {steps.map((step, index) => {
              const isSelected = step.id === selected
              return (
                <button
                  key={step.id}
                  type="button"
                  aria-pressed={isSelected}
                  aria-label={`Paso ${index + 1}: ${step.name}. ${step.kicker}. ${step.detail}`}
                  onClick={() => onSelect(step.id)}
                  style={{
                    position: 'absolute',
                    left: layout.centerX - layout.nodeWidth / 2,
                    top: layout.nodeTop(index),
                    width: layout.nodeWidth,
                    height: layout.nodeHeight,
                    padding: 0,
                    textAlign: 'left',
                    cursor: 'pointer',
                    overflow: 'hidden',
                    font: 'inherit',
                    color: 'var(--mantine-color-black)',
                    background: 'var(--mantine-color-white)',
                    borderRadius: 'var(--mantine-radius-md)',
                    border: isSelected
                      ? `2px solid ${color.selected}`
                      : `1px solid ${step.problem ? 'var(--mantine-color-red-3)' : 'var(--mantine-color-gray-3)'}`,
                    boxShadow: isSelected
                      ? '0 0 0 4px var(--mantine-color-fiuba-1), var(--mantine-shadow-md)'
                      : 'var(--mantine-shadow-sm)',
                  }}
                >
                  <Text
                    component="span"
                    size="xs"
                    fw={700}
                    tt="uppercase"
                    display="flex"
                    style={{
                      justifyContent: 'space-between',
                      padding: '3px 12px',
                      background: step.problem
                        ? 'var(--mantine-color-red-0)'
                        : step.applicant
                          ? 'var(--mantine-color-fiuba-0)'
                          : 'var(--mantine-color-gray-1)',
                      color: step.problem
                        ? color.problem
                        : step.applicant
                          ? 'var(--mantine-color-fiuba-8)'
                          : 'var(--mantine-color-gray-8)',
                    }}
                  >
                    <span>{step.kicker}</span>
                    <span>{index + 1}</span>
                  </Text>
                  <Box px="sm" py={4}>
                    <Text component="span" display="block" fw={600} truncate="end">
                      {step.name}
                    </Text>
                    <Text component="span" display="block" size="xs" c="dimmed" truncate="end">
                      {step.detail}
                    </Text>
                  </Box>
                </button>
              )
            })}
          </div>
          {toolbar && (
            <Group justify="center" mt="sm">
              {toolbar}
            </Group>
          )}
        </Box>
        <Paper shadow="md" radius="md" p="lg" withBorder style={{ flex: '0 1 340px', minWidth: 280 }}>
          {panel}
        </Paper>
      </Group>
    </Paper>
  )
}

// terminal places the start, filled, or the close, outlined, on the center
// line of the diagram.
function terminal(centerX: number, top: number, filled: boolean) {
  return {
    position: 'absolute' as const,
    left: centerX,
    top,
    transform: 'translateX(-50%)',
    padding: '4px 12px',
    borderRadius: 999,
    background: filled ? 'var(--mantine-color-black)' : 'var(--mantine-color-white)',
    border: '2px solid var(--mantine-color-black)',
  }
}
