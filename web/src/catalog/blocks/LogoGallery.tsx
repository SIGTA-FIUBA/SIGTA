import { Anchor, Box, Paper, SimpleGrid, Stack, Text, Title } from '@mantine/core'
import { Frame } from './Frame'

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
  { id: 'black', name: 'Negro', background: '#ffffff', use: 'Impresión en un color.' },
  {
    id: 'white',
    name: 'Blanco',
    background: '#121516',
    use: 'Sobre fotos o fondos oscuros: todo blanco, también la barra.',
  },
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
            <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="sm">
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
