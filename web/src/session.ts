import { redirect } from 'react-router'
import { api } from './api/client'

// requireSession guards the pages that need a session: without one, the
// browser goes to the login page. It returns the person with the session.
export async function requireSession() {
  const { data, response } = await api.GET('/me')
  if (response.status === 401) {
    throw redirect('/login')
  }
  if (!data) {
    throw new Error(`GET /me answered ${response.status}`)
  }
  return data
}

// skipLoginWithSession sends whoever already has a session to the home page.
export async function skipLoginWithSession() {
  const { response } = await api.GET('/me')
  if (response.ok) {
    throw redirect('/')
  }
  return null
}

// logout closes the session and reports whether the API did.
export async function logout(): Promise<boolean> {
  const { response } = await api.POST('/auth/logout')
  return response.ok
}
