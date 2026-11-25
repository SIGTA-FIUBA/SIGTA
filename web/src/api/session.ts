import { redirect, useSearchParams, type LoaderFunctionArgs } from 'react-router'
import { api } from './client'

// requireSession guards the pages that need a session: without one, the
// browser goes to the login page, which logs in back to the page it asked for
// (useLoginUrl). It returns the person with the session.
export async function requireSession({ request }: LoaderFunctionArgs) {
  const user = await optionalSession()
  if (user === null) {
    const { pathname, search } = new URL(request.url)
    const page = pathname + search
    throw redirect(page === '/' ? '/login' : `/login?${new URLSearchParams({ return_to: page })}`)
  }
  return user
}

// optionalSession is for the pages anyone can read, such as the catalog: it
// returns the person with the session, or null for a visitor.
export async function optionalSession() {
  const { data, response } = await api.GET('/me')
  if (response.status === 401) {
    return null
  }
  if (!data) {
    throw new Error(`GET /me answered ${response.status}`)
  }
  return data
}

// publicSession is the session of the pages anyone can read: the landing, the
// catalog and the page about SIGTA. If the API cannot say who is there, they
// still open, for a visitor. Whoever has a session goes from the landing to the
// home page.
export async function publicSession({ request }: { request: Request }) {
  const user = await optionalSession().catch(() => null)
  if (user !== null && new URL(request.url).pathname === '/login') {
    throw redirect('/')
  }
  return user
}

// loginUrl is where the login starts: a page of the backend, so its link
// loads it instead of routing.
export const loginUrl = '/api/auth/login'

// loginUrlTo is the login that comes back to a page of the application, such
// as the form of the procedure a visitor asked to start.
export function loginUrlTo(page: string) {
  return `${loginUrl}?${new URLSearchParams({ return_to: page })}`
}

// useLoginUrl is the login of the public pages: back to the page that sent the
// person to log in (requireSession), if one did. The backend checks that it is
// a page of SIGTA.
export function useLoginUrl() {
  const [params] = useSearchParams()
  const page = params.get('return_to')
  return page ? loginUrlTo(page) : loginUrl
}

// logout closes the session and reports whether the API did.
export async function logout(): Promise<boolean> {
  const { response } = await api.POST('/auth/logout')
  return response.ok
}
