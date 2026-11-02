import { redirect } from 'react-router'
import { api } from './api/client'

// requireSession guards the pages that need a session: without one, the
// browser goes to the login page.
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
