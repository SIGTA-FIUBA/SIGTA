import type { ReactNode } from 'react'

const shift = { start: undefined, middle: 'translateX(-50%)', end: 'translateX(-100%)' }

// Placed puts what it holds at a point of the diagram, in pixels computed by
// stepLayout. It is the one element of the interface styled inline (ADR-015,
// web/DESIGN.md): a position is a coordinate, not a design value, and the
// theme has no measure for it. Everything it holds takes its look from the
// theme. A frame is the box the rest is placed in.
export function Placed({
  x = 0,
  y = 0,
  width,
  height,
  maxWidth,
  anchor = 'start',
  frame = false,
  children,
}: {
  x?: number
  y?: number
  width?: number
  height?: number
  maxWidth?: number
  anchor?: keyof typeof shift
  frame?: boolean
  children: ReactNode
}) {
  return (
    <div
      style={
        frame
          ? { position: 'relative', width, height, margin: '0 auto' }
          : { position: 'absolute', left: x, top: y, width, height, maxWidth, transform: shift[anchor] }
      }
    >
      {children}
    </div>
  )
}
