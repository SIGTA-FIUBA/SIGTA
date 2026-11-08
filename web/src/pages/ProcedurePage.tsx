import { Anchor, Stack, Text, Title } from '@mantine/core'
import { Link, useLocation } from 'react-router'

// ProcedurePage is where the inbox leads. The detail of a procedure comes with
// its own story (US19); for now the page only leads back to the list with the
// filters it had.
export function ProcedurePage() {
  const location = useLocation()
  const state: unknown = location.state
  const from =
    typeof state === 'object' && state !== null && 'from' in state && typeof state.from === 'string'
      ? state.from
      : '/inbox'

  return (
    <Stack gap="md">
      <Title order={1}>Trámite</Title>
      <Text c="dimmed">El detalle del trámite todavía no está disponible.</Text>
      <Anchor component={Link} to={from}>
        Volver a la bandeja
      </Anchor>
    </Stack>
  )
}
