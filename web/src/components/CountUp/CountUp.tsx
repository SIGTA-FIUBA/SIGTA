import { VisuallyHidden } from '@mantine/core'
import { animate, useReducedMotion } from 'motion/react'
import { useEffect, useRef, useState } from 'react'
import { count } from '../../theme/motion'

// CountUp writes a figure that counts up to its value when it arrives, and
// from the old value to the new one when it changes, in motion.count: the
// change is seen, not only its result. A screen reader reads the value
// itself, never the numbers on the way. With reduced motion it shows the
// value at once.
export function CountUp({ value }: { value: number }) {
  const reduce = useReducedMotion() === true
  const [shown, setShown] = useState(0)
  const from = useRef(0)

  useEffect(() => {
    if (reduce) {
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
  }, [value, reduce])

  return (
    <>
      <VisuallyHidden>{value}</VisuallyHidden>
      <span aria-hidden>{reduce ? value : shown}</span>
    </>
  )
}
