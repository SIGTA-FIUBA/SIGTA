import { useSearchParams } from 'react-router'
import type { components } from '../api/schema'
import './LoginPage.css'

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

// LoginPage leaves for the identity provider with a full navigation, since the
// provider sends the browser back to the API's callback.
export function LoginPage() {
  const [params] = useSearchParams()
  const message = messageFor(params.get('error'))

  return (
    <main className="login">
      <h1>SIGTA</h1>
      <p>Sistema Integral de Gestión de Trámites Académicos</p>
      {message && (
        <p className="login-error" role="alert">
          {message}
        </p>
      )}
      <a className="google-button" href="/api/auth/login">
        <GoogleLogo />
        <span>Iniciar sesión con Google</span>
      </a>
    </main>
  )
}

function GoogleLogo() {
  return (
    <svg viewBox="0 0 48 48" width="20" height="20" aria-hidden="true">
      <path
        fill="#EA4335"
        d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"
      />
      <path
        fill="#4285F4"
        d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"
      />
      <path
        fill="#FBBC05"
        d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"
      />
      <path
        fill="#34A853"
        d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"
      />
    </svg>
  )
}
