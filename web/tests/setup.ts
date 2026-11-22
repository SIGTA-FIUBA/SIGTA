import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { MotionGlobalConfig } from 'motion/react'
import { afterEach, vi } from 'vitest'

afterEach(() => {
  cleanup()
})

// The tests check what a screen shows and does, not its motion: every
// animation ends at once, so nothing waits at opacity 0.
MotionGlobalConfig.skipAnimations = true

// jsdom lacks the browser APIs that Mantine components read, as listed in
// https://mantine.dev/guides/vitest.
const { getComputedStyle } = window
window.getComputedStyle = (element) => getComputedStyle(element)
window.HTMLElement.prototype.scrollIntoView = () => {}

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

window.ResizeObserver = class {
  observe() {}
  unobserve() {}
  disconnect() {}
}

// jsdom has no layout, so every element counts as in view: what enters when
// it is scrolled to (Reveal inView, CountUp) shows at once.
window.IntersectionObserver = class {
  readonly root = null
  readonly rootMargin = ''
  readonly scrollMargin = ''
  readonly thresholds = []
  readonly callback: IntersectionObserverCallback
  constructor(callback: IntersectionObserverCallback) {
    this.callback = callback
  }
  observe(target: Element) {
    queueMicrotask(() =>
      this.callback([{ target, isIntersecting: true, intersectionRatio: 1 } as IntersectionObserverEntry], this),
    )
  }
  unobserve() {}
  disconnect() {}
  takeRecords() {
    return []
  }
}

// Textarea with autosize measures again once the fonts load.
Object.defineProperty(document, 'fonts', {
  value: { addEventListener: vi.fn(), removeEventListener: vi.fn(), ready: Promise.resolve() },
})
