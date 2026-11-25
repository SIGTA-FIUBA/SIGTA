import { Text, Title } from '@mantine/core'

// The text styles of tokens.json, with the component that writes each one.
export const textStyles = [
  { name: 'title-page', use: 'Title order={1}', sample: (t: string) => <Title order={1}>{t}</Title> },
  { name: 'title-section', use: 'Title order={2}', sample: (t: string) => <Title order={2}>{t}</Title> },
  { name: 'title-item', use: 'Title order={3}, Text fw={600}', sample: (t: string) => <Text fw={600}>{t}</Text> },
  { name: 'body', use: 'Text', sample: (t: string) => <Text>{t}</Text> },
  { name: 'body-sm', use: 'Text size="sm"', sample: (t: string) => <Text size="sm">{t}</Text> },
  {
    name: 'label',
    use: 'Text size="sm" fw={500}',
    sample: (t: string) => (
      <Text size="sm" fw={500}>
        {t}
      </Text>
    ),
  },
  { name: 'caption', use: 'Text size="xs"', sample: (t: string) => <Text size="xs">{t}</Text> },
]
