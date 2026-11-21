import { Spoiler, Text } from '@mantine/core'

// LongValue keeps the lines the applicant wrote and folds a long text.
export function LongValue({ text }: { text: string }) {
  return (
    <Spoiler showLabel="Ver todo" hideLabel="Ver menos">
      {text.split('\n').map((line, i) => (
        <Text key={i}>{line === '' ? ' ' : line}</Text>
      ))}
    </Spoiler>
  )
}
