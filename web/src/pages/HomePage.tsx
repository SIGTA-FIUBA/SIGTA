import { Stack, Text, Title } from '@mantine/core'

export function HomePage() {
  return (
    <Stack gap="xs">
      <Title order={1}>Inicio</Title>
      <Text c="dimmed">Sistema Integral de Gestión de Trámites Académicos</Text>
    </Stack>
  )
}
