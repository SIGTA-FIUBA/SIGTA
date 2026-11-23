import type { RouteObject } from 'react-router'
import { AppLayout } from './layout/AppLayout'
import { AboutPage } from './pages/AboutPage'
import { AdministrationPage } from './pages/AdministrationPage'
import { CatalogPage } from './pages/CatalogPage'
import { HomePage } from './pages/HomePage'
import { InboxPage } from './pages/InboxPage'
import { LoginPage } from './pages/LoginPage'
import { MyProceduresPage } from './pages/MyProceduresPage'
import { ProcedurePage } from './pages/ProcedurePage'
import { ProcedureTypeDraftPage } from './pages/ProcedureTypeDraftPage'
import { ProcedureTypesPage } from './pages/ProcedureTypesPage'
import { StartProcedurePage } from './pages/StartProcedurePage'
import { ProcedurePreview } from './dev/ProcedurePreview'
import { StartProcedurePreview } from './dev/StartProcedurePreview'
import { AdministrationPreview } from './dev/AdministrationPreview'
import { HomePreview } from './dev/HomePreview'
import { ProcedureTypeDraftPreview } from './dev/ProcedureTypeDraftPreview'
import { optionalSession, requireSession, skipLoginWithSession } from './api/session'

export const routes: RouteObject[] = [
  // The landing, the catalog and the page about SIGTA ask for no session, so
  // they can be read before logging in. A visitor reads them under one header
  // that stays from one to another; whoever has a session gets the catalog
  // and the page about SIGTA with the menu, and the home page instead of the
  // landing.
  {
    loader: optionalSession,
    element: <AppLayout />,
    children: [
      { path: '/login', loader: skipLoginWithSession, element: <LoginPage /> },
      { path: '/catalog', element: <CatalogPage /> },
      { path: '/about', element: <AboutPage /> },
    ],
  },
  // Only the dev server has the preview with worst-case data.
  ...(import.meta.env.DEV
    ? [
        { path: '/dev/start-procedure/:procedureTypeId', element: <StartProcedurePreview /> },
        { path: '/dev/procedures/:id', element: <ProcedurePreview /> },
        { path: '/dev/administration', element: <AdministrationPreview /> },
        { path: '/dev/procedure-types/:procedureTypeId/draft', element: <ProcedureTypeDraftPreview /> },
        { path: '/dev/home', element: <HomePreview />, children: [{ index: true, element: <HomePage /> }] },
      ]
    : []),
  {
    loader: requireSession,
    element: <AppLayout />,
    children: [
      { path: '/', element: <HomePage /> },
      { path: '/catalog/:procedureTypeId/start', element: <StartProcedurePage /> },
      { path: '/my-procedures', element: <MyProceduresPage /> },
      { path: '/inbox', element: <InboxPage /> },
      { path: '/procedures/:id', element: <ProcedurePage /> },
      { path: '/administration', element: <AdministrationPage /> },
      { path: '/procedure-types', element: <ProcedureTypesPage /> },
      { path: '/procedure-types/:procedureTypeId/draft', element: <ProcedureTypeDraftPage /> },
    ],
  },
]
