import { Affix, Container, SegmentedControl } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { ProcedurePage } from '../pages/ProcedurePage'
import { fixtures, type FixtureName } from './procedureFixtures'

const fixtureOptions: { value: FixtureName; label: string }[] = [
  { value: 'demo', label: 'Demo' },
  { value: 'worst', label: 'Worst case' },
  { value: 'empty', label: 'Empty' },
  { value: 'one', label: 'One' },
]

function isFixture(value: string | null): value is FixtureName {
  return value !== null && value in fixtures
}

// ProcedurePreview renders the page of a procedure with each detail of
// procedureFixtures, swapped in the query cache, which is where the page reads
// it. Only for development: the route exists only in the dev server. The
// fixture stays in ?data= across reloads.
export function ProcedurePreview() {
  const [params, setParams] = useSearchParams()
  const fixtureParam = params.get('data')
  const fixture: FixtureName = isFixture(fixtureParam) ? fixtureParam : 'demo'
  const client = useMemo(() => {
    const seeded = new QueryClient({ defaultOptions: { queries: { staleTime: Infinity, retry: false } } })
    seeded.setQueryData(['procedure', fixtures[fixture].id], fixtures[fixture])
    return seeded
  }, [fixture])

  return (
    <>
      <QueryClientProvider client={client} key={fixture}>
        <Container size="md" py="md">
          <ProcedurePage />
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
