import {
  Anchor,
  Badge,
  Box,
  Code,
  Group,
  MantineProvider,
  Paper,
  SimpleGrid,
  Stack,
  Table,
  Text,
  Title,
} from '@mantine/core'
import { Unstyled } from '@storybook/addon-docs/blocks'
import type { ReactNode } from 'react'
import { cssVariablesResolver, theme } from '../theme/theme'
import { tokenSource, tokens, value } from '../theme/tokens'

// Blocks of the catalog's documentation pages (the .mdx files). They draw the
// tokens from tokens.json with the app's theme, so the pages cannot drift from
// the code. Only the catalog uses them; showing a raw value is their purpose,
// which is why .oxlintrc.json lets this folder write styles.

const SIZES = ['xs', 'sm', 'md', 'lg', 'xl'] as const
const SPACING = ['xxs', 'xs', 'sm', 'md', 'lg', 'xl', 'xxl', 'xxxl'] as const

// Frame puts a block inside the app's theme, away from the styles of the
// documentation page.
export function Frame({ children }: { children: ReactNode }) {
  return (
    <Unstyled>
      <MantineProvider theme={theme} cssVariablesResolver={cssVariablesResolver} forceColorScheme="light">
        <Box py="sm">{children}</Box>
      </MantineProvider>
    </Unstyled>
  )
}

function luminance(hex: string) {
  const channels = [1, 3, 5].map((start) => Number.parseInt(hex.slice(start, start + 2), 16) / 255)
  const [r, g, b] = channels.map((c) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4))
  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

// contrast is the WCAG contrast ratio between two colors written as #rrggbb.
function contrast(a: string, b: string) {
  const [light, dark] = [luminance(a), luminance(b)].sort((x, y) => y - x)
  return (light + 0.05) / (dark + 0.05)
}

function Swatch({ color, label, note }: { color: string; label: string; note?: string }) {
  const onWhite = contrast(color, '#ffffff')
  const ink = onWhite >= 4.5 ? '#ffffff' : tokens.color.ink[900]
  return (
    <Paper withBorder radius="sm" style={{ overflow: 'hidden' }}>
      <Box p="sm" style={{ background: color, color: ink, minHeight: 72 }}>
        <Text fw={700} size="sm">
          {label}
        </Text>
        <Text size="xs" ff="monospace">
          {color.toLowerCase()}
        </Text>
      </Box>
      <Stack gap={4} p="xs">
        <Text size="xs">
          {onWhite.toFixed(2)}:1 sobre blanco{' '}
          {onWhite >= 4.5 ? (
            <Badge size="xs">texto</Badge>
          ) : onWhite >= 3 ? (
            <Badge size="xs" color="gray">
              gráficos
            </Badge>
          ) : null}
        </Text>
        {note && (
          <Text size="xs" c="dimmed">
            {note}
          </Text>
        )}
      </Stack>
    </Paper>
  )
}

const SCALE_NOTES: Record<string, Record<number, string>> = {
  fiuba: {
    0: 'brand.50: selección, avatar',
    5: 'brand.500: cian institucional',
    7: 'action.600: primario',
    8: 'action.700: hover',
  },
  gray: {
    0: 'surface.page: fondo',
    1: 'surface.hover',
    3: 'border.subtle: bordes y divisores',
    4: 'border.control: campos',
    6: 'ink.500: metadatos',
    7: 'ink.600: texto atenuado (dimmed)',
    9: 'ink.900: texto',
  },
  green: { 0: 'success.50', 7: 'success.700' },
  orange: { 0: 'warning.50', 7: 'warning.700' },
  red: { 0: 'danger.50', 7: 'danger.700' },
}

// ColorScale shows the ten shades of a Mantine color, with their contrast on
// white and the semantic token each one carries.
export function ColorScale({ name }: { name: keyof typeof tokens.color.scales }) {
  const notes = SCALE_NOTES[name] ?? {}
  return (
    <Frame>
      <SimpleGrid cols={{ base: 2, sm: 5 }} spacing="xs">
        {tokens.color.scales[name].map((shade, index) => (
          <Swatch key={shade} color={shade} label={`${name}.${index}`} note={notes[index]} />
        ))}
      </SimpleGrid>
    </Frame>
  )
}

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

// FontFamilies shows each family with a sample.
export function FontFamilies() {
  const families = tokenSource.font.family
  const rows = [
    { name: 'sans', stack: tokens.font.sans, note: families.sans.$description },
    { name: 'mono', stack: tokens.font.mono, note: families.mono.$description },
    { name: 'brand', stack: tokens.font.brand, note: families.brand.$description },
  ]
  return (
    <Frame>
      <Stack gap="sm">
        {rows.map((row) => (
          <Paper key={row.name} withBorder p="md">
            <Text size="xl" style={{ fontFamily: row.name === 'brand' ? undefined : row.stack }}>
              {row.name === 'brand' ? 'Montserrat, dibujada en el logo' : 'Equivalencia entre asignaturas 0123456789'}
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

const TEXT_STYLES = [
  { name: 'title-page', use: 'Title order={1}', sample: (t: string) => <Title order={1}>{t}</Title> },
  { name: 'title-section', use: 'Title order={2}', sample: (t: string) => <Title order={2}>{t}</Title> },
  { name: 'title-item', use: 'Title order={3}, Text fw={700}', sample: (t: string) => <Text fw={700}>{t}</Text> },
  { name: 'body', use: 'Text', sample: (t: string) => <Text>{t}</Text> },
  { name: 'body-sm', use: 'Text size="sm"', sample: (t: string) => <Text size="sm">{t}</Text> },
  { name: 'caption', use: 'Text size="xs"', sample: (t: string) => <Text size="xs">{t}</Text> },
]

// TextStyles shows the six text styles, the only ones, as the screens write
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
          {TEXT_STYLES.map((style) => (
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

// Headings shows Title order 1 to 3 and the display variant of the landing
// page. Orders 4 to 6 look like order 3.
export function Headings() {
  const orders = [1, 2, 3] as const
  return (
    <Frame>
      <Stack gap="md">
        <Box>
          <Text size="xs" c="dimmed">
            Title variant=&quot;display&quot;: crece con el ancho de pantalla, solo en la pantalla de ingreso
          </Text>
          <Title order={1} variant="display">
            Trámites académicos, sin papeles
          </Title>
        </Box>
        {orders.map((order) => (
          <Box key={order}>
            <Text size="xs" c="dimmed">
              Title order={order}: {TEXT_STYLES[order - 1]!.name}
            </Text>
            <Title order={order}>Equivalencia entre asignaturas</Title>
          </Box>
        ))}
      </Stack>
    </Frame>
  )
}

// Spacing shows the spacing scale as bars.
export function Spacing() {
  return (
    <Frame>
      <Stack gap="xs">
        {SPACING.map((size) => (
          <Group key={size} gap="md" wrap="nowrap">
            <Code w={48}>{size}</Code>
            <Text size="sm" w={80}>
              {tokenSource.spacing[size].$value}
            </Text>
            <Box h={16} bg="fiuba.5" style={{ width: `var(--mantine-spacing-${size})` }} />
          </Group>
        ))}
      </Stack>
    </Frame>
  )
}

// Radii shows each radius on a panel and marks the default one.
export function Radii() {
  return (
    <Frame>
      <SimpleGrid cols={{ base: 2, sm: 5 }} spacing="sm">
        {SIZES.map((size) => (
          <Paper key={size} withBorder radius={size} p="md" bg="fiuba.0">
            <Group gap="xs">
              <Code>{size}</Code>
              {size === tokens.defaultRadius && <Badge size="xs">por defecto</Badge>}
            </Group>
            <Text size="xs" mt="xs">
              {tokenSource.radius[size].$value}
            </Text>
          </Paper>
        ))}
      </SimpleGrid>
    </Frame>
  )
}

// Shadows shows the only shadow, on what floats.
export function Shadows() {
  return (
    <Frame>
      <Paper shadow="md" p="md" maw={240}>
        <Text size="sm">Un menú desplegado</Text>
      </Paper>
    </Frame>
  )
}

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

const PIECES = [
  { id: 'isotipo', name: 'Isotipo', use: 'Solo el edificio. Favicon, avatar, espacios chicos.' },
  { id: 'logotipo', name: 'Logotipo', use: 'Solo la palabra. Cuando el isotipo ya está cerca.' },
  { id: 'horizontal', name: 'Isologo horizontal', use: 'El de uso general: encabezados y documentos.' },
  { id: 'horizontal-descriptor', name: 'Isologo horizontal con descriptor', use: 'Presentaciones y portadas.' },
  { id: 'vertical', name: 'Isologo vertical', use: 'Espacios cuadrados o angostos.' },
  { id: 'vertical-descriptor', name: 'Isologo vertical con descriptor', use: 'Portadas y piezas impresas.' },
]

const VARIANTS = [
  { id: 'color', name: 'Color', background: '#ffffff', use: 'Sobre blanco o fondos claros.' },
  { id: 'color-dark', name: 'Color para fondo oscuro', background: '#003A6A', use: 'Sobre azul oscuro o negro.' },
  { id: 'black', name: 'Negro', background: '#ffffff', use: 'Impresión en un color.' },
  { id: 'white', name: 'Blanco', background: '#121516', use: 'Sobre fotos o colores oscuros, en un color.' },
]

// LogoGallery shows every piece of the logo in every variant, with the file
// to download.
export function LogoGallery({ piece }: { piece?: string }) {
  const pieces = piece ? PIECES.filter((p) => p.id === piece) : PIECES
  return (
    <Frame>
      <Stack gap="xl">
        {pieces.map((p) => (
          <Stack key={p.id} gap="xs">
            <Title order={3}>{p.name}</Title>
            <Text size="sm" c="dimmed">
              {p.use}
            </Text>
            <SimpleGrid cols={{ base: 1, sm: 2, md: 4 }} spacing="sm">
              {VARIANTS.map((v) => {
                const file = `brand/sigta-${p.id}-${v.id}.svg`
                return (
                  <Paper key={v.id} withBorder radius="sm" style={{ overflow: 'hidden' }}>
                    <Box
                      p="lg"
                      h={160}
                      style={{
                        background: v.background,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                      }}
                    >
                      <img
                        src={file}
                        alt={`${p.name}, ${v.name.toLowerCase()}`}
                        style={{ maxWidth: '100%', maxHeight: '100%' }}
                      />
                    </Box>
                    <Stack gap={2} p="xs">
                      <Text size="sm" fw={700}>
                        {v.name}
                      </Text>
                      <Text size="xs" c="dimmed">
                        {v.use}
                      </Text>
                      <Anchor href={file} download size="xs" ff="monospace">
                        {file.replace('brand/', '')}
                      </Anchor>
                    </Stack>
                  </Paper>
                )
              })}
            </SimpleGrid>
          </Stack>
        ))}
      </Stack>
    </Frame>
  )
}
