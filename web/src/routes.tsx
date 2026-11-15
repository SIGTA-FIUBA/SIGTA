import type { RouteObject } from 'react-router'
import { AppLayout } from './layout/AppLayout'
import { AboutPage } from './pages/AboutPage'
import { CatalogPage } from './pages/CatalogPage'
import { HomePage } from './pages/HomePage'
import { InboxPage } from './pages/InboxPage'
import { LoginPage } from './pages/LoginPage'
import { ProcedurePage } from './pages/ProcedurePage'
import { optionalSession, requireSession, skipLoginWithSession } from './session'

export const routes: RouteObject[] = [
  { path: '/login', loader: skipLoginWithSession, element: <LoginPage /> },
  // The catalog and the page about SIGTA ask for no session, so they can be
  // read before logging in. They keep the same frame, with the menu for
  // whoever has a session.
  {
    loader: optionalSession,
    element: <AppLayout />,
    children: [
      { path: '/catalog', element: <CatalogPage /> },
      { path: '/about', element: <AboutPage /> },
    ],
  },
  {
    loader: requireSession,
    element: <AppLayout />,
    children: [
      { path: '/', element: <HomePage /> },
      { path: '/inbox', element: <InboxPage /> },
      { path: '/procedures/:id', element: <ProcedurePage /> },
    ],
  },
]
