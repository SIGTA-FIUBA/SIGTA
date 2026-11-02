import { createBrowserRouter, RouterProvider } from 'react-router'
import { CatalogPage } from './pages/CatalogPage'
import { HomePage } from './pages/HomePage'

const router = createBrowserRouter([
  { path: '/', element: <HomePage /> },
  { path: '/catalog', element: <CatalogPage /> },
])

export function App() {
  return <RouterProvider router={router} />
}
