import { Anchor, Stack, Text, Title } from '@mantine/core'
import { Link, useOutletContext } from 'react-router'
import type { components } from '../api/schema'

type User = components['schemas']['User']

export function HomePage() {
  // The layout passes the person with the session; outside it there is none.
  const user = useOutletContext<User | null | undefined>()
  const staff = user != null && user.role !== 'applicant'

  return (
    <Stack gap="xs">
      <Title order={1}>Inicio</Title>
      <Text c="dimmed">Sistema Integral de Gestión de Trámites Académicos</Text>
      {staff && (
        <Anchor component={Link} to="/inbox">
          Ver la bandeja
        </Anchor>
      )}
      <Anchor component={Link} to="/catalog">
        Ver los trámites disponibles
      </Anchor>
    </Stack>
  )
}
