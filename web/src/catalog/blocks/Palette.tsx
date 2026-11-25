import { Code, SimpleGrid, Stack, Text } from '@mantine/core'
import { tokenSource, value } from '../../theme/tokens'
import { Frame } from './Frame'
import { Swatch } from './Swatch'

type ColorNode = Record<string, { $value?: string; $description?: string }> & { $description?: string }

// Palette shows the semantic colors of a group of tokens.json, such as
// "brand" or "ink", with what each one is for.
export function Palette({ groups }: { groups: string[] }) {
  return (
    <Frame>
      <Stack gap="md">
        {groups.map((group) => {
          const node = (tokenSource.color as unknown as Record<string, ColorNode>)[group]!
          const keys = Object.keys(node).filter((key) => !key.startsWith('$'))
          return (
            <Stack key={group} gap="xs">
              <Text size="sm">
                <Code>color.{group}</Code> {node.$description}
              </Text>
              <SimpleGrid cols={{ base: 2, sm: 4 }} spacing="xs">
                {keys.map((key) => (
                  <Swatch
                    key={key}
                    color={String(value(`color.${group}.${key}`))}
                    label={`${group}.${key}`}
                    note={node[key]!.$description}
                  />
                ))}
              </SimpleGrid>
            </Stack>
          )
        })}
      </Stack>
    </Frame>
  )
}
