import { Code, Table } from '@mantine/core'
import { value } from '../../theme/tokens'
import { Frame } from './Frame'
import { textStyles } from './textStyleList'

// TextStyles shows the seven text styles, the only ones, as the screens write
// them.
export function TextStyles() {
  return (
    <Frame>
      <Table>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Estilo</Table.Th>
            <Table.Th>Medida</Table.Th>
            <Table.Th>En las pantallas</Table.Th>
            <Table.Th>Ejemplo</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {textStyles.map((style) => (
            <Table.Tr key={style.name}>
              <Table.Td>
                <Code>{style.name}</Code>
              </Table.Td>
              <Table.Td>
                {String(value(`textStyle.${style.name}.size`))} / {String(value(`textStyle.${style.name}.lineHeight`))},{' '}
                {String(value(`textStyle.${style.name}.weight`))}
              </Table.Td>
              <Table.Td>
                <Code>{style.use}</Code>
              </Table.Td>
              <Table.Td>{style.sample('Equivalencia entre asignaturas')}</Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </Frame>
  )
}
