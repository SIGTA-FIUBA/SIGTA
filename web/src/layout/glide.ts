import type { MouseEvent } from 'react'

// A plain click: one that opens a link in the same tab, which the page can
// answer itself. With a modifier or another button the browser does its own.
export const isPlainClick = (event: MouseEvent) =>
  event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey

// glide brings into sight what a link of the same page points to, or the
// top of the page: gliding, or at once with reduced motion. The target's
// scroll-margin-top leaves room for the header that stays on top.
export function glide(target: Element | 'top' | null) {
  const behavior = window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth'
  if (target === 'top') {
    window.scrollTo({ top: 0, behavior })
  } else {
    target?.scrollIntoView({ behavior, block: 'start' })
  }
}
