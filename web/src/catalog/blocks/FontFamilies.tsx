import { Code, Paper, Stack, Text } from '@mantine/core'
import { tokenSource, tokens } from '../../theme/tokens'
import { Frame } from './Frame'

// FontFamilies shows each family with a sample.
export function FontFamilies() {
  const families = tokenSource.font.family
  const rows = [
    { name: 'sans', stack: tokens.font.sans, note: families.sans.$description },
    { name: 'mono', stack: tokens.font.mono, note: families.mono.$description },
    { name: 'brand', stack: tokens.font.brand, note: families.brand.$description },
    { name: 'logo', stack: tokens.font.logo, note: families.logo.$description },
    { name: 'landing', stack: tokens.font.landing, note: families.landing.$description },
  ]
  return (
    <Frame>
      <Stack gap="sm">
        {rows.map((row) => (
          <Paper key={row.name} withBorder p="md">
            <Text size="xl" style={{ fontFamily: row.name === 'logo' ? undefined : row.stack }}>
              {row.name === 'logo' ? 'Montserrat, dibujada en el logo' : 'Equivalencia entre asignaturas 0123456789'}
            </Text>
            <Text size="sm" mt="xs">
              <Code>font.family.{row.name}</Code> {row.note}
            </Text>
            <Text size="xs" c="dimmed" ff="monospace">
              {row.stack}
            </Text>
          </Paper>
        ))}
      </Stack>
    </Frame>
  )
}
