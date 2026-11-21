import { Code, Table, Text } from '@mantine/core'
import { tokenSource } from '../../theme/tokens'
import { Frame } from './Frame'
import { SIZES } from './catalogValues'

// TextSizes shows the body sizes, xs to xl, as Text renders them.
export function TextSizes() {
  return (
    <Frame>
      <Table>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Tamaño</Table.Th>
            <Table.Th>Medida</Table.Th>
            <Table.Th>Interlineado</Table.Th>
            <Table.Th>Ejemplo</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {SIZES.map((size) => (
            <Table.Tr key={size}>
              <Table.Td>
                <Code>{size}</Code>
              </Table.Td>
              <Table.Td>{tokenSource.font.size[size].$value}</Table.Td>
              <Table.Td>{tokenSource.font.lineHeight[size].$value}</Table.Td>
              <Table.Td>
                <Text size={size}>El trámite espera la firma del titular.</Text>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </Frame>
  )
}
