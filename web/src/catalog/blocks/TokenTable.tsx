import { Code, Table, Text } from '@mantine/core'
import { tokenSource } from '../../theme/tokens'
import { Frame } from './Frame'

type TokenNode = { $value?: unknown; $description?: string; [key: string]: unknown }

function flatten(node: TokenNode, path: string): { path: string; value: unknown; description?: string }[] {
  if ('$value' in node) {
    return [{ path, value: node.$value, description: node.$description }]
  }
  return Object.entries(node)
    .filter(([key]) => !key.startsWith('$'))
    .flatMap(([key, child]) => flatten(child as TokenNode, path ? `${path}.${key}` : key))
}

// TokenTable lists every token under a path of tokens.json, with its value and
// what it is for.
export function TokenTable({ path }: { path: string }) {
  let node = tokenSource as unknown as TokenNode
  for (const key of path.split('.')) {
    node = node[key] as TokenNode
  }
  return (
    <Frame>
      <Table>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Token</Table.Th>
            <Table.Th>Valor</Table.Th>
            <Table.Th>Uso</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {flatten(node, path).map((token) => (
            <Table.Tr key={token.path}>
              <Table.Td>
                <Code>{token.path}</Code>
              </Table.Td>
              <Table.Td>
                <Code>{Array.isArray(token.value) ? token.value.join(', ') : String(token.value)}</Code>
              </Table.Td>
              <Table.Td>
                <Text size="sm">{token.description}</Text>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </Frame>
  )
}
