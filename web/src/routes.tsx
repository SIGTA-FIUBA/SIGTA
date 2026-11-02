import type { RouteObject } from 'react-router'
import { AppLayout } from './layout/AppLayout'
import { CatalogPage } from './pages/CatalogPage'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/LoginPage'
import { requireSession, skipLoginWithSession } from './session'

export const routes: RouteObject[] = [
  { path: '/login', loader: skipLoginWithSession, element: <LoginPage /> },
  // The catalog asks for no session, so it can be read before logging in.
  { path: '/catalog', element: <CatalogPage /> },
  {
    loader: requireSession,
    element: <AppLayout />,
    children: [{ path: '/', element: <HomePage /> }],
  },
]
