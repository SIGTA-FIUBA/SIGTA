import { Greeting } from '../features/greeting/Greeting'

export function HomePage() {
  return (
    <main>
      <h1>SIGTA</h1>
      <p>Sistema Integral de Gestión de Trámites Académicos</p>
      <Greeting />
    </main>
  )
}
