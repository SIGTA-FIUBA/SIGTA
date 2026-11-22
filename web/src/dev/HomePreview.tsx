import { Affix, Container, SegmentedControl } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useMemo } from 'react'
import { Outlet, useSearchParams } from 'react-router'
import { fixtures, type FixtureName } from './homeFixtures'

const fixtureOptions: { value: FixtureName; label: string }[] = [
  { value: 'demo', label: 'Demo' },
  { value: 'worst', label: 'Worst case' },
  { value: 'empty', label: 'Empty' },
  { value: 'one', label: 'One' },
  { value: 'many', label: '40 closed' },
]

const user = { id: 'u1', email: 'titular@example.test', name: 'Titular', staff: false, administrator: false }

function isFixture(value: string | null): value is FixtureName {
  return value !== null && value in fixtures
}

// HomePreview renders the home page with each list of homeFixtures, swapped in
// the query cache, which is where the page reads it. Only for development: the
// route exists only in the dev server. The fixture stays in ?data= across
// reloads.
export function HomePreview() {
  const [params, setParams] = useSearchParams()
  const fixtureParam = params.get('data')
  const fixture: FixtureName = isFixture(fixtureParam) ? fixtureParam : 'demo'
  const client = useMemo(() => {
    const seeded = new QueryClient({ defaultOptions: { queries: { staleTime: Infinity, retry: false } } })
    seeded.setQueryData(['me', 'procedures'], fixtures[fixture])
    return seeded
  }, [fixture])

  return (
    <>
      <QueryClientProvider client={client} key={fixture}>
        <Container size="lg" py="md">
          <Outlet context={user} />
        </Container>
      </QueryClientProvider>
      <Affix position={{ bottom: 'md', left: '50%' }}>
        <SegmentedControl
          size="xs"
          value={fixture}
          data={fixtureOptions}
          onChange={(value) => setParams({ data: value }, { replace: true })}
        />
      </Affix>
    </>
  )
}
