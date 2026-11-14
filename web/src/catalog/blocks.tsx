import {
  Anchor,
  Badge,
  Box,
  Code,
  DEFAULT_THEME,
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
import { tokenSource, tokens } from '../theme/tokens'

// Blocks of the catalog's documentation pages (the .mdx files). They draw the
// tokens from tokens.json with the app's theme, so the pages cannot drift from
// the code. Only the catalog uses them; showing a raw value is their purpose,
// which is why .oxlintrc.json lets this folder write styles.

const SIZES = ['xs', 'sm', 'md', 'lg', 'xl'] as const

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
  const ink = onWhite >= 4.5 ? '#ffffff' : tokens.color.text
  return (
    <Paper withBorder radius="sm" style={{ overflow: 'hidden' }}>
      <Box p="sm" style={{ background: color, color: ink, minHeight: 72 }}>
        <Text fw={600} size="sm">
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

// ColorScale shows the ten shades of a Mantine color, with their contrast on white.
export function ColorScale({ name }: { name: string }) {
  const shades = name === 'fiuba' ? tokens.color.fiuba : DEFAULT_THEME.colors[name]
  const notes: Record<number, string> =
    name === 'fiuba'
      ? { 5: 'Celeste de la marca, fondos', 7: 'Primario: botones y links', 8: 'Hover del primario' }
      : name === 'gray'
        ? { 0: 'Fondo de las páginas', 7: 'Texto atenuado (dimmed)' }
        : {}
  return (
    <Frame>
      <SimpleGrid cols={{ base: 2, sm: 5 }} spacing="xs">
        {shades.map((shade, index) => (
          <Swatch key={shade} color={shade} label={`${name}.${index}`} note={notes[index]} />
        ))}
      </SimpleGrid>
    </Frame>
  )
}

// BrandColors shows the colors of the logo and of the text.
export function BrandColors() {
  const brand = tokenSource.color.brand
  return (
    <Frame>
      <SimpleGrid cols={{ base: 1, sm: 4 }} spacing="xs">
        <Swatch color={brand.navy.$value} label="brand.navy" note={brand.navy.$description} />
        <Swatch color={brand.cyan.$value} label="brand.cyan" note={brand.cyan.$description} />
        <Swatch color={brand.ink.$value} label="brand.ink" note={brand.ink.$description} />
        <Swatch color={tokens.color.text} label="text" note="Texto de la interfaz (Mantine black)." />
      </SimpleGrid>
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

// Headings shows Title order 1 to 6 and the display variant.
export function Headings() {
  const orders = [1, 2, 3, 4, 5, 6] as const
  return (
    <Frame>
      <Stack gap="md">
        <Box>
          <Text size="xs" c="dimmed">
            Title variant=&quot;display&quot;: crece con el ancho de pantalla, solo en la portada
          </Text>
          <Title order={1} variant="display">
            Trámites académicos, sin papeles
          </Title>
        </Box>
        {orders.map((order) => {
          const heading = tokenSource.heading[`h${order}`]
          return (
            <Box key={order}>
              <Text size="xs" c="dimmed">
                Title order={order}: {heading.fontSize.$value}, interlineado {heading.lineHeight.$value}, peso{' '}
                {tokens.font.headingWeight}
              </Text>
              <Title order={order}>Equivalencia entre asignaturas</Title>
            </Box>
          )
        })}
      </Stack>
    </Frame>
  )
}

// Spacing shows the spacing scale as bars.
export function Spacing() {
  return (
    <Frame>
      <Stack gap="xs">
        {SIZES.map((size) => (
          <Group key={size} gap="md" wrap="nowrap">
            <Code w={40}>{size}</Code>
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

// Shadows shows each shadow on a panel.
export function Shadows() {
  return (
    <Frame>
      <SimpleGrid cols={{ base: 2, sm: 5 }} spacing="lg">
        {SIZES.map((size) => (
          <Paper key={size} shadow={size} p="md">
            <Code>{size}</Code>
          </Paper>
        ))}
      </SimpleGrid>
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
                      <Text size="sm" fw={600}>
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
