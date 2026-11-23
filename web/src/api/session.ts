import { redirect } from 'react-router'
import { api } from './client'

// requireSession guards the pages that need a session: without one, the
// browser goes to the login page. It returns the person with the session.
export async function requireSession() {
  const user = await optionalSession()
  if (user === null) {
    throw redirect('/login')
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

// logout closes the session and reports whether the API did.
export async function logout(): Promise<boolean> {
  const { response } = await api.POST('/auth/logout')
  return response.ok
}
