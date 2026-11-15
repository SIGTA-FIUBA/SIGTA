import { Box, Group, VisuallyHidden } from '@mantine/core'

// StepBar shows where a procedure is among the steps of its type: one segment
// per step, the done ones in action.600, the open one in brand.500 and the
// ones to come in border.subtle. Read aloud, it says "Etapa n de m".
export function StepBar({ step, count }: { step: number; count: number }) {
  return (
    <>
      <Group aria-hidden gap="var(--sigta-step-bar-gap)" wrap="nowrap" w="100%">
        {Array.from({ length: count }, (_, index) => (
          <Box
            key={index}
            flex={1}
            h="var(--sigta-step-bar-height)"
            bdrs="xl"
            bg={index + 1 < step ? 'fiuba.7' : index + 1 === step ? 'fiuba.5' : 'gray.3'}
          />
        ))}
      </Group>
      <VisuallyHidden>{`Etapa ${step} de ${count}`}</VisuallyHidden>
    </>
  )
}
