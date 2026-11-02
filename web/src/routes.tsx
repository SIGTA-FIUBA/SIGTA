import type { RouteObject } from 'react-router'
import { AppLayout } from './layout/AppLayout'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/LoginPage'
import { requireSession, skipLoginWithSession } from './session'

export const routes: RouteObject[] = [
  { path: '/login', loader: skipLoginWithSession, element: <LoginPage /> },
  {
    loader: requireSession,
    element: <AppLayout />,
    children: [{ path: '/', element: <HomePage /> }],
  },
]
