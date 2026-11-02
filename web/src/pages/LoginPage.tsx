import { Alert, Container, Stack, Text, Title } from '@mantine/core'
import { useSearchParams } from 'react-router'
import type { components } from '../api/schema'
import { GoogleSignInButton } from '../components/GoogleSignInButton/GoogleSignInButton'

type LoginError = components['schemas']['LoginError']

// The callback only says why a login failed; the text lives here, so the
// language of the interface does not depend on the backend.
const messages: Record<LoginError, string> = {
  state_mismatch: 'El ingreso no coincide con el que se inició. Volvé a intentarlo.',
  rejected: 'El ingreso no se pudo completar. Volvé a intentarlo.',
  domain_not_allowed: 'Ingresá con tu cuenta institucional.',
}

function messageFor(error: string | null): string | undefined {
  if (error === null) {
    return undefined
  }
  return Object.hasOwn(messages, error) ? messages[error as LoginError] : messages.rejected
}

export function LoginPage() {
  const [params] = useSearchParams()
  const message = messageFor(params.get('error'))

  return (
    <Container size="xs">
      <Stack align="center" justify="center" gap="xl" mih="100dvh">
        <Stack align="center" gap="xs">
          <Title order={1}>SIGTA</Title>
          <Text c="dimmed" ta="center">
            Sistema Integral de Gestión de Trámites Académicos
          </Text>
        </Stack>
        {message && (
          <Alert color="red" w="100%">
            {message}
          </Alert>
        )}
        <GoogleSignInButton />
      </Stack>
    </Container>
  )
}
