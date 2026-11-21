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
})
