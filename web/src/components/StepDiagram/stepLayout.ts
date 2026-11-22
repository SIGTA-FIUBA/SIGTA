// stepLayout places the steps of a procedure type from top to bottom, in the
// order of the definition, and routes a line for every way out of each one:
// down to the next step, around the right side to a later one, and around the
// left side, dashed, to an earlier one (a send back or any line that goes
// back). It is plain geometry, in pixels, so the diagram only draws it.
//
// The width leaves room for every curve and label, so nothing falls outside
// the diagram. A compact layout, for a narrow screen, draws tighter curves and
// names only the lines going down, on the line itself: the names of the side
// lines are in the panel of the step.

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
  // labelAnchor says which point of the label sits at labelX.
  labelAnchor: 'start' | 'middle' | 'end'
}

export type Layout = {
  width: number
  height: number
  centerX: number
  nodeWidth: number
  nodeHeight: number
  // labelWidth is the most a label takes; a longer one is cut.
  labelWidth: number
  nodeTop: (index: number) => number
  closeTop: number
  edges: LayoutEdge[]
}

const nodeHeight = 108
const pitch = 170
const firstTop = 70
const gap = 8

export function stepLayout(steps: LayoutStep[], { compact = false } = {}): Layout {
  const nodeWidth = compact ? 200 : 220
  const curve = compact ? 30 : 60
  const laneGap = compact ? 12 : 26
  const labelWidth = compact ? 120 : 200
  const count = steps.length
  const indexOf = new Map(steps.map((step, index) => [step.id, index]))
  const top = (index: number) => (index === count ? firstTop + count * pitch - 20 : firstTop + index * pitch)
  const middle = (index: number) => (index === count ? top(count) + 15 : top(index) + nodeHeight / 2)
  const target = (id: string) => indexOf.get(id) ?? count

  type Exit = { from: number; to: number; label?: string; sendBack: boolean; bend: number; nth: number }
  let rightLane = 0
  let leftLane = 0
  const exits = steps.flatMap<Exit>((step, index) =>
    (step.outcomes && step.outcomes.length > 0
      ? step.outcomes.map((outcome) => ({ to: target(outcome.next), label: outcome.name, sendBack: outcome.send_back }))
      : [{ to: target(step.next ?? ''), label: undefined, sendBack: false }]
    ).map((exit, nth) => {
      const bend =
        exit.to === index + 1 ? 0 : exit.to > index ? curve + laneGap * rightLane++ : curve + laneGap * leftLane++
      return { ...exit, from: index, bend, nth }
    }),
  )
  const sideLabels = !compact
  const reach = (exit: Exit) => 0.75 * exit.bend + (sideLabels && exit.label ? gap + labelWidth : 0)
  const leftNeed = Math.max(40, ...exits.filter((exit) => exit.to < exit.from).map(reach)) + gap
  const downLabels = exits.some((exit) => exit.to === exit.from + 1 && exit.label)
  const rightNeed =
    Math.max(
      40,
      !compact && downLabels ? 14 + labelWidth : 0,
      ...exits.filter((exit) => exit.to > exit.from + 1).map(reach),
    ) + gap
  const centerX = nodeWidth / 2 + leftNeed
  const width = centerX + nodeWidth / 2 + rightNeed
  const right = centerX + nodeWidth / 2
  const left = centerX - nodeWidth / 2

  // Several outcomes of a step can lead to the next one: their labels go one
  // under the other, centered between both steps.
  const downs = exits.filter((exit) => exit.to === exit.from + 1)
  const downCount = (from: number) => downs.filter((exit) => exit.from === from).length
  const downIndex = (exit: Exit) => downs.filter((other) => other.from === exit.from).indexOf(exit)
  const labelHeight = 22

  const edges: LayoutEdge[] = exits.map((exit) => {
    const from = steps[exit.from].id
    const shared = { from, sendBack: exit.sendBack }
    if (exit.to === exit.from + 1) {
      const y1 = top(exit.from) + nodeHeight
      const y2 = top(exit.to) - 2
      return {
        ...shared,
        kind: 'down',
        label: exit.label,
        path: `M ${centerX} ${y1} L ${centerX} ${y2}`,
        head: `M ${centerX} ${y2} L ${centerX - 5} ${y2 - 9} L ${centerX + 5} ${y2 - 9} Z`,
        labelX: compact ? centerX : centerX + 14,
        labelY: (y1 + y2) / 2 - (downCount(exit.from) * labelHeight) / 2 + downIndex(exit) * labelHeight,
        labelAnchor: compact ? 'middle' : 'start',
      }
    }
    const ya = middle(exit.from)
    const yb = middle(exit.to)
    if (exit.to > exit.from) {
      const xb = (exit.to === count ? centerX + 42 : right) + 2
      // Beside the step it leaves, one under the other, so as not to meet the
      // labels of the lines going down, which sit between the steps.
      return {
        ...shared,
        kind: 'forward',
        label: sideLabels ? exit.label : undefined,
        path: `M ${right} ${ya} C ${right + exit.bend} ${ya} ${right + exit.bend} ${yb} ${xb} ${yb}`,
        head: `M ${xb} ${yb} L ${xb + 9} ${yb - 5} L ${xb + 9} ${yb + 5} Z`,
        labelX: right + 0.75 * exit.bend + gap,
        labelY: top(exit.from) + 4 + 26 * exit.nth,
        labelAnchor: 'start',
      }
    }
    const xb = left - 2
    return {
      ...shared,
      kind: 'back',
      label: sideLabels ? exit.label : undefined,
      path: `M ${left} ${ya} C ${left - exit.bend} ${ya} ${left - exit.bend} ${yb} ${xb} ${yb}`,
      head: `M ${xb} ${yb} L ${xb - 9} ${yb - 5} L ${xb - 9} ${yb + 5} Z`,
      labelX: left - 0.75 * exit.bend - gap,
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
    labelWidth,
    nodeTop: top,
    closeTop: top(count),
    edges,
  }
}
