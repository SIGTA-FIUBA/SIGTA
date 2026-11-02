import { Link } from 'react-router'

export function HomePage() {
  return (
    <main>
      <h1>SIGTA</h1>
      <p>Sistema Integral de Gestión de Trámites Académicos</p>
      <p>
        <Link to="/catalog">Ver los trámites disponibles</Link>
      </p>
    </main>
  )
}
