import { useGreeting } from './useGreeting'

export function Greeting() {
  const { data, isPending, isError } = useGreeting()

  if (isPending) return <p>Cargando…</p>
  if (isError) return <p role="alert">No se pudo contactar al backend.</p>

  return <p>{data.message}</p>
}
