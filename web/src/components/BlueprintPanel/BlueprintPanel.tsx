import { Box, Paper, type PaperProps } from '@mantine/core'
import type { PointerEvent, ReactNode } from 'react'

type BlueprintPanelProps = PaperProps & {
  // component is the element of the panel, such as section or header.
  component?: 'div' | 'section' | 'header'
  'aria-labelledby'?: string
  children: ReactNode
}

// BlueprintPanel is the navy panel with the technical grid of a blueprint
// (Paper variant="blueprint"): the header of the home page and of the
// landing page. Under a mouse the grid lights up around the pointer, which
// says the panel is alive without asking for attention; a finger or a pen
// leaves it still. The text on it is white or fiuba.1 and keeps its contrast
// without the texture.
export function BlueprintPanel({ component = 'div', children, ...props }: BlueprintPanelProps) {
  function follow(event: PointerEvent<HTMLDivElement>) {
    if (event.pointerType !== 'mouse') {
      return
    }
    const panel = event.currentTarget
    const box = panel.getBoundingClientRect()
    panel.style.setProperty('--pointer-x', `${event.clientX - box.left}px`)
    panel.style.setProperty('--pointer-y', `${event.clientY - box.top}px`)
    panel.dataset.pointer = ''
  }

  function leave(event: PointerEvent<HTMLDivElement>) {
    delete event.currentTarget.dataset.pointer
  }

  return (
    <Paper {...props} component={component} variant="blueprint" onPointerMove={follow} onPointerLeave={leave}>
      <Box data-grain aria-hidden />
      {children}
    </Paper>
  )
}
