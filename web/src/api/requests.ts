import { ForbiddenError } from './procedures'

// ActionError carries the API's message, which says why a change was refused
// and what to do.
class ActionError extends Error {}

const connectionFailure = 'No pudimos guardar el cambio. Revisá tu conexión y volvé a intentarlo.'
const unknownFailure = 'No pudimos guardar el cambio. Volvé a intentarlo.'

// change makes a change and returns what the API answered, or throws
// ActionError with the reason it was refused.
export async function change<T>(
  call: Promise<{ data?: T; error?: unknown; response: Response }>,
): Promise<T | undefined> {
  let outcome: { data?: T; error?: unknown; response: Response }
  try {
    outcome = await call
  } catch {
    throw new ActionError(connectionFailure)
  }
  if (!outcome.response.ok) {
    const { error } = outcome
    const message =
      typeof error === 'object' && error !== null && 'message' in error && typeof error.message === 'string'
        ? error.message
        : unknownFailure
    throw new ActionError(message)
  }
  return outcome.data
}

// forbiddenOr returns what a read answered, or throws ForbiddenError when the
// person cannot see it.
export function forbiddenOr<T>(path: string, data: T | undefined, response: Response): T {
  if (response.status === 403) {
    throw new ForbiddenError()
  }
  if (data === undefined) {
    throw new Error(`GET ${path} answered ${response.status}`)
  }
  return data
}

// run makes a change whose answer is not needed.
export async function run(call: Promise<{ error?: unknown; response: Response }>): Promise<void> {
  await change(call)
}
