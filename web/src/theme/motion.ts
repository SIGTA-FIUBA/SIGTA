import type { Transition, Variants } from 'motion/react'
import { tokens } from './tokens'

// The movement of the interface, for the components that animate with Motion
// (motion.dev). Like the rest of the theme, every duration, distance and curve
// comes from tokens.json. Motion follows prefers-reduced-motion (MotionConfig
// in main.tsx): with it, nothing moves and what arrives only fades in.

const seconds = (ms: number) => ms / 1000

// enter is what arrives with the page: it rises motion.rise and fades in, in
// motion.slow. The custom value is its position in a list, which delays it
// motion.stagger after the one before, up to motion.staggerLimit.
export const enter: Variants = {
  hidden: { opacity: 0, y: tokens.motion.rise },
  shown: (index: number = 0) => ({
    opacity: 1,
    y: 0,
    transition: {
      duration: seconds(tokens.motion.slow),
      ease: tokens.motion.easingOut,
      delay: seconds(tokens.motion.stagger) * Math.min(index, tokens.motion.staggerLimit),
    },
  }),
}

// move is how something that changes place travels there: the mark of a
// selected tile, a row that the new order puts elsewhere.
export const move: Transition = { duration: seconds(tokens.motion.slow), ease: tokens.motion.easingOut }

// count is how a figure counts up to its value.
export const count: Transition = { duration: seconds(tokens.motion.count), ease: tokens.motion.easingOut }

// viewport is when something below the first screen counts as arrived: once,
// when a fifth of it is in view, so it is never seen waiting.
export const viewport = { once: true, amount: 0.2 } as const

// swap is how one page of the public frame gives way to the next one under
// the same header: the one leaving fades and rises a little, out of focus,
// and the one arriving comes into focus from below.
export const swap: Variants = {
  hidden: { opacity: 0, y: tokens.motion.rise, filter: `blur(${tokens.motion.blur}px)` },
  shown: {
    opacity: 1,
    y: 0,
    filter: 'blur(0px)',
    transition: { duration: seconds(tokens.motion.slow), ease: tokens.motion.easingOut },
  },
  gone: {
    opacity: 0,
    y: -tokens.motion.rise / 2,
    filter: `blur(${tokens.motion.blur}px)`,
    transition: { duration: seconds(tokens.motion.fast), ease: tokens.motion.easingOut },
  },
}
