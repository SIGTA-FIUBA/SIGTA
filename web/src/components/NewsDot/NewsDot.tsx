import { Box, VisuallyHidden } from '@mantine/core'

// NewsDot marks, next to a title, a procedure that moved since the person last
// saw it: a brand.500 dot of 8. A screen reader hears "Con novedades".
export function NewsDot() {
  return (
    <>
      <Box
        component="span"
        aria-hidden
        display="inline-block"
        w="var(--sigta-news-dot)"
        h="var(--sigta-news-dot)"
        bdrs="xl"
        bg="fiuba.5"
        flex="none"
      />
      <VisuallyHidden>Con novedades</VisuallyHidden>
    </>
  )
}
