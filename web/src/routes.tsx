import type { RouteObject } from 'react-router'
import { AppLayout } from './layout/AppLayout'
import { AdministrationPage } from './pages/AdministrationPage'
import { CatalogPage } from './pages/CatalogPage'
import { HomePage } from './pages/HomePage'
import { InboxPage } from './pages/InboxPage'
import { LoginPage } from './pages/LoginPage'
import { ProcedurePage } from './pages/ProcedurePage'
import { StartProcedurePage } from './pages/StartProcedurePage'
import { ProcedurePreview } from './dev/ProcedurePreview'
import { StartProcedurePreview } from './dev/StartProcedurePreview'
import { AdministrationPreview } from './dev/AdministrationPreview'
import { optionalSession, requireSession, skipLoginWithSession } from './api/session'

export const routes: RouteObject[] = [
  { path: '/login', loader: skipLoginWithSession, element: <LoginPage /> },
  // The catalog asks for no session, so it can be read before logging in. It
  // keeps the same frame, with the menu for whoever has a session.
  { loader: optionalSession, element: <AppLayout />, children: [{ path: '/catalog', element: <CatalogPage /> }] },
  // Only the dev server has the preview with worst-case data.
  ...(import.meta.env.DEV
    ? [
        { path: '/dev/start-procedure/:procedureTypeId', element: <StartProcedurePreview /> },
        { path: '/dev/procedures/:id', element: <ProcedurePreview /> },
        { path: '/dev/administration', element: <AdministrationPreview /> },
      ]
    : []),
  {
    loader: requireSession,
    element: <AppLayout />,
    children: [
      { path: '/', element: <HomePage /> },
      { path: '/catalog/:procedureTypeId/start', element: <StartProcedurePage /> },
      { path: '/inbox', element: <InboxPage /> },
      { path: '/procedures/:id', element: <ProcedurePage /> },
      { path: '/administration', element: <AdministrationPage /> },
    ],
  },
]
