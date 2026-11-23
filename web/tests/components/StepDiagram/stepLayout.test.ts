import { describe, expect, it } from 'vitest'
import { stepLayout } from '../../../src/components/StepDiagram/stepLayout'

const steps = [
  { id: 'request', next: 'review' },
  {
    id: 'review',
    outcomes: [
      { name: 'Continuar', next: 'resolution', send_back: false },
      { name: 'Pedir información', next: 'request', send_back: true },
      { name: 'Archivar', next: '', send_back: false },
    ],
  },
  { id: 'resolution' },
]

describe('stepLayout', () => {
  it('routes each way out of a step down, forward on the right or back on the left', () => {
    const layout = stepLayout(steps)
    const [start, ...exits] = layout.edges

    expect(start.from).toBe('')
    expect(exits.map((edge) => [edge.from, edge.kind, edge.label])).toEqual([
      ['request', 'down', undefined],
      ['review', 'down', 'Continuar'],
      ['review', 'back', 'Pedir información'],
      ['review', 'forward', 'Archivar'],
      ['resolution', 'down', undefined],
    ])
    expect(exits[2].sendBack).toBe(true)
    expect(exits[2].labelAnchor).toBe('end')
  })

  it('places the steps from top to bottom and the close after the last one', () => {
    const layout = stepLayout(steps)

    expect(layout.nodeTop(1)).toBeGreaterThan(layout.nodeTop(0) + layout.nodeHeight)
    expect(layout.closeTop).toBeGreaterThan(layout.nodeTop(2) + layout.nodeHeight)
    expect(layout.height).toBeGreaterThan(layout.closeTop)
  })

  it('draws no line into an empty route', () => {
    expect(stepLayout([]).edges).toEqual([])
  })

  it('stacks the labels of outcomes that lead to the same next step', () => {
    const layout = stepLayout([
      {
        id: 'review',
        outcomes: [
          { name: 'Aprobar', next: 'close', send_back: false },
          { name: 'Archivar', next: 'close', send_back: false },
        ],
      },
      { id: 'close' },
    ])
    const [approve, archive] = layout.edges.filter((edge) => edge.label)

    expect(archive.labelY - approve.labelY).toBeGreaterThanOrEqual(20)
  })

  it('goes compact on a narrow screen, naming only the lines going down', () => {
    const steps = [
      { id: 'a', next: 'b' },
      {
        id: 'b',
        outcomes: [
          { name: 'Volver', next: 'a', send_back: true },
          { name: 'Seguir', next: 'c', send_back: false },
        ],
      },
      { id: 'c' },
    ]
    const wide = stepLayout(steps)
    const narrow = stepLayout(steps, { compact: true })

    expect(narrow.width).toBeLessThan(330)
    expect(narrow.edges.find((edge) => edge.kind === 'back')?.label).toBeUndefined()
    expect(wide.edges.find((edge) => edge.kind === 'back')?.label).toBe('Volver')
    expect(narrow.edges.find((edge) => edge.label === 'Seguir')?.labelAnchor).toBe('middle')
  })

  it('leaves room for every label inside the diagram', () => {
    const layout = stepLayout(steps)

    for (const edge of layout.edges.filter((edge) => edge.label)) {
      const left = edge.labelAnchor === 'end' ? edge.labelX - layout.labelWidth : edge.labelX
      expect(left).toBeGreaterThanOrEqual(0)
      expect(left + layout.labelWidth).toBeLessThanOrEqual(layout.width)
    }
  })
})
