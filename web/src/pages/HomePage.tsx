import { Anchor, Stack, Text, Title } from '@mantine/core'
import { Link } from 'react-router'

export function HomePage() {
  return (
    <Stack gap="xs">
      <Title order={1}>Inicio</Title>
      <Text c="dimmed">Sistema Integral de Gestión de Trámites Académicos</Text>
      <Anchor component={Link} to="/catalog">
        Ver los trámites disponibles
      </Anchor>
    </Stack>
  )
}
