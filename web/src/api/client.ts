import createClient from 'openapi-fetch'
import type { paths } from './schema'

// Typed client of the public API, generated from contracts/openapi.yaml. The
// reverse proxy serves the API under /api of the same origin. fetch is looked
// up on each request, not captured once, so tests can stub it.
export const api = createClient<paths>({
  baseUrl: `${window.location.origin}/api`,
  fetch: (request) => globalThis.fetch(request),
})
