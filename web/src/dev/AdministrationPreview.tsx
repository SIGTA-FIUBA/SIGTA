import { Affix, Container, SegmentedControl } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { AdministrationPage } from '../pages/AdministrationPage'
import { fixtures, type FixtureName } from './administrationFixtures'

const fixtureOptions: { value: FixtureName; label: string }[] = [
  { value: 'demo', label: 'Demo' },
  { value: 'worst', label: 'Worst case' },
  { value: 'empty', label: 'Empty' },
]

function isFixture(value: string | null): value is FixtureName {
  return value !== null && value in fixtures
}

// AdministrationPreview renders the administration with each fixture of
// administrationFixtures, swapped in the query cache, which is where the page
// reads it. Only for development: the route exists only in the dev server.
// Changes still go to the API. The fixture stays in ?data= across reloads.
export function AdministrationPreview() {
  const [params, setParams] = useSearchParams()
  const fixtureParam = params.get('data')
  const fixture: FixtureName = isFixture(fixtureParam) ? fixtureParam : 'demo'
  const client = useMemo(() => {
    const data = fixtures[fixture]
    const seeded = new QueryClient({ defaultOptions: { queries: { staleTime: Infinity, retry: false } } })
    seeded.setQueryData(['administration', 'areas'], data.areas)
    for (const [area, memberships] of Object.entries(data.memberships)) {
      seeded.setQueryData(['administration', 'areas', area, 'memberships'], memberships)
    }
    seeded.setQueryData(['administration', 'administrators'], data.administrators)
    seeded.setQueryData(['procedure-types'], data.procedureTypes)
    return seeded
  }, [fixture])

  return (
    <>
      <QueryClientProvider client={client} key={fixture}>
        <Container size="xl" py="md">
          <AdministrationPage />
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
