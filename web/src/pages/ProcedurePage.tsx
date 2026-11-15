import { Anchor, Stack, Text, Title } from '@mantine/core'
import { useEffect } from 'react'
import { Link, useLocation, useParams } from 'react-router'
import { useMarkSeen } from '../api/procedures'

// ProcedurePage is where the inbox leads. The detail of a procedure comes with
// its own story (US19); for now the page only leads back to the list with the
// filters it had. Opening it counts as seeing the procedure, so the lists stop
// marking it as news.
export function ProcedurePage() {
  const location = useLocation()
  const { id } = useParams()
  const { mutate: markSeen } = useMarkSeen()
  useEffect(() => {
    if (id !== undefined) {
      markSeen(id)
    }
  }, [id, markSeen])
  const state: unknown = location.state
  const from =
    typeof state === 'object' && state !== null && 'from' in state && typeof state.from === 'string'
      ? state.from
      : '/inbox'

  return (
    <Stack gap="md">
      <Title order={1}>Trámite</Title>
      <Text c="dimmed">El detalle del trámite estará disponible próximamente.</Text>
      <Anchor component={Link} to={from}>
        Volver a la bandeja de trámites
      </Anchor>
    </Stack>
  )
}
