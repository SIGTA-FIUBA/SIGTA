import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Greeting } from './Greeting'

function renderGreeting() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <Greeting />
    </QueryClientProvider>,
  )
}

describe('Greeting', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('shows the message from the backend', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json({ message: 'Hello world!' })),
    )

    renderGreeting()

    expect(await screen.findByText('Hello world!')).toBeInTheDocument()
  })

  it('reports when the backend answers with an error', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(null, { status: 500 })),
    )

    renderGreeting()

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudo contactar al backend.')
  })

  it('reports when the backend cannot be reached', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => {
        throw new TypeError('Failed to fetch')
      }),
    )

    renderGreeting()

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudo contactar al backend.')
  })
})
