import { Stack, Text, Title } from '@mantine/core'

export function HomePage() {
  return (
    <Stack gap="xs" p="md">
      <Title order={1}>SIGTA</Title>
      <Text c="dimmed">Sistema Integral de Gestión de Trámites Académicos</Text>
    </Stack>
  )
}
