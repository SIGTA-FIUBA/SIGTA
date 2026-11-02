import type { RouteObject } from 'react-router'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/LoginPage'
import { requireSession, skipLoginWithSession } from './session'

export const routes: RouteObject[] = [
  { path: '/login', loader: skipLoginWithSession, element: <LoginPage /> },
  {
    loader: requireSession,
    children: [{ path: '/', element: <HomePage /> }],
  },
]
