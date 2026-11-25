import { VisuallyHidden } from '@mantine/core'
import { animate, useInView, useReducedMotion } from 'motion/react'
import { useEffect, useRef, useState } from 'react'
import { count, viewport } from '../../theme/motion'

// CountUp writes a figure that counts up to its value when it is in view, and
// from the old value to the new one when it changes, in motion.count: the
// change is seen, not only its result. A screen reader reads the value
// itself, never the numbers on the way. With reduced motion it shows the
// value at once.
export function CountUp({ value }: { value: number }) {
  const reduce = useReducedMotion() === true
  const [shown, setShown] = useState(0)
  const from = useRef(0)
  const figure = useRef<HTMLSpanElement>(null)
  const seen = useInView(figure, viewport)

  useEffect(() => {
    if (reduce || !seen) {
      return
    }
    const controls = animate(from.current, value, {
      ...count,
      onUpdate: (latest) => {
        from.current = latest
        setShown(Math.round(latest))
      },
    })
    return () => controls.stop()
  }, [value, reduce, seen])

  return (
    <>
      <VisuallyHidden>{value}</VisuallyHidden>
      <span ref={figure} aria-hidden>
        {reduce ? value : shown}
      </span>
    </>
  )
}
