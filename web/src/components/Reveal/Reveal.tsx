import { motion } from 'motion/react'
import type { ReactNode } from 'react'
import { enter } from '../../theme/motion'

const elements = { div: motion.div, li: motion.li, section: motion.section }

type RevealProps = {
  // index is the position in a list: each one arrives a little after the one
  // before, up to motion.staggerLimit.
  index?: number
  // as is the element that holds the content, li inside a list.
  as?: keyof typeof elements
  children: ReactNode
}

// Reveal brings its content in when the page arrives: it rises a little and
// fades in, once. It orients: the page builds up from the top, in the order
// in which it is read. With reduced motion the content only fades in.
export function Reveal({ index = 0, as = 'div', children }: RevealProps) {
  const Element = elements[as]
  return (
    <Element variants={enter} initial="hidden" animate="shown" custom={index}>
      {children}
    </Element>
  )
}
