// stepLayout places the steps of a procedure type from top to bottom, in the
// order of the definition, and routes a line for every way out of each one:
// down to the next step, around the right side to a later one, and around the
// left side, dashed, to an earlier one (a send back or any line that goes
// back). It is plain geometry, in pixels, so the diagram only draws it.

export type LayoutStep = {
  id: string
  // next is where a step that does not decide goes; empty closes the procedure.
  next?: string
  // outcomes are the ways out of a step that decides.
  outcomes?: { name: string; next: string; send_back: boolean }[]
}

type LayoutEdge = {
  from: string
  path: string
  head: string
  kind: 'down' | 'forward' | 'back'
  sendBack: boolean
  label?: string
  labelX: number
  labelY: number
  // labelAnchor says which side of the label sits at labelX.
  labelAnchor: 'start' | 'end'
}

export type Layout = {
  width: number
  height: number
  centerX: number
  nodeWidth: number
  nodeHeight: number
  nodeTop: (index: number) => number
  closeTop: number
  edges: LayoutEdge[]
}

const nodeWidth = 220
const nodeHeight = 108
const pitch = 170
const firstTop = 70
const centerX = 330
const width = 640
const curve = 60
const laneGap = 26

export function stepLayout(steps: LayoutStep[]): Layout {
  const count = steps.length
  const indexOf = new Map(steps.map((step, index) => [step.id, index]))
  const top = (index: number) => (index === count ? firstTop + count * pitch - 20 : firstTop + index * pitch)
  const middle = (index: number) => (index === count ? top(count) + 15 : top(index) + nodeHeight / 2)
  const right = centerX + nodeWidth / 2
  const left = centerX - nodeWidth / 2
  const target = (id: string) => indexOf.get(id) ?? count

  type Exit = { from: number; to: number; label?: string; sendBack: boolean }
  const exits: Exit[] = steps.flatMap<Exit>((step, index) =>
    step.outcomes && step.outcomes.length > 0
      ? step.outcomes.map((outcome) => ({
          from: index,
          to: target(outcome.next),
          label: outcome.name,
          sendBack: outcome.send_back,
        }))
      : [{ from: index, to: target(step.next ?? ''), label: undefined, sendBack: false }],
  )

  let rightLane = 0
  let leftLane = 0
  const edges: LayoutEdge[] = exits.map((exit) => {
    const from = steps[exit.from].id
    if (exit.to === exit.from + 1) {
      const y1 = top(exit.from) + nodeHeight
      const y2 = top(exit.to) - 2
      return {
        from,
        kind: 'down',
        sendBack: exit.sendBack,
        label: exit.label,
        path: `M ${centerX} ${y1} L ${centerX} ${y2}`,
        head: `M ${centerX} ${y2} L ${centerX - 5} ${y2 - 9} L ${centerX + 5} ${y2 - 9} Z`,
        labelX: centerX + 14,
        labelY: (y1 + y2) / 2 - 10,
        labelAnchor: 'start',
      }
    }
    if (exit.to > exit.from) {
      const bend = curve + laneGap * rightLane++
      const ya = middle(exit.from)
      const yb = middle(exit.to)
      const xb = (exit.to === count ? centerX + 42 : right) + 2
      return {
        from,
        kind: 'forward',
        sendBack: exit.sendBack,
        label: exit.label,
        path: `M ${right} ${ya} C ${right + bend} ${ya} ${right + bend} ${yb} ${xb} ${yb}`,
        head: `M ${xb} ${yb} L ${xb + 9} ${yb - 5} L ${xb + 9} ${yb + 5} Z`,
        labelX: right + 0.75 * bend + 8,
        labelY: (ya + yb) / 2 - 10,
        labelAnchor: 'start',
      }
    }
    const bend = curve + laneGap * leftLane++
    const ya = middle(exit.from)
    const yb = middle(exit.to)
    const xb = left - 2
    return {
      from,
      kind: 'back',
      sendBack: exit.sendBack,
      label: exit.label,
      path: `M ${left} ${ya} C ${left - bend} ${ya} ${left - bend} ${yb} ${xb} ${yb}`,
      head: `M ${xb} ${yb} L ${xb - 9} ${yb - 5} L ${xb - 9} ${yb + 5} Z`,
      labelX: left - 0.75 * bend - 8,
      labelY: (ya + yb) / 2 - 10,
      labelAnchor: 'end',
    }
  })

  // The line from the start into the first step.
  if (count > 0) {
    edges.unshift({
      from: '',
      kind: 'down',
      sendBack: false,
      path: `M ${centerX} 42 L ${centerX} ${firstTop - 2}`,
      head: `M ${centerX} ${firstTop - 2} L ${centerX - 5} ${firstTop - 11} L ${centerX + 5} ${firstTop - 11} Z`,
      labelX: centerX,
      labelY: 0,
      labelAnchor: 'start',
    })
  }

  return {
    width,
    height: top(count) + 80,
    centerX,
    nodeWidth,
    nodeHeight,
    nodeTop: top,
    closeTop: top(count),
    edges,
  }
}
