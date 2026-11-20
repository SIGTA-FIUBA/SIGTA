import { Affix, Container, SegmentedControl } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { ProcedurePage } from '../pages/ProcedurePage'
import { fixtures, type FixtureName } from './procedureFixtures'

const states: { value: FixtureName; label: string }[] = [
  { value: 'demo', label: 'Demo' },
  { value: 'worst', label: 'Worst case' },
  { value: 'empty', label: 'Empty' },
  { value: 'one', label: 'One' },
  { value: 'deciding', label: 'Deciding' },
  { value: 'sentBack', label: 'Sent back' },
  { value: 'sentBackWorst', label: 'Sent back, worst case' },
]

function isFixture(value: string | null): value is FixtureName {
  return value !== null && value in fixtures
}

// ProcedurePreview renders the page of a procedure with each detail of
// procedureFixtures, swapped in the query cache, which is where the page reads
// it. Only for development: the route exists only in the dev server. The state
// stays in ?data= across reloads.
export function ProcedurePreview() {
  const [params, setParams] = useSearchParams()
  const data = params.get('data')
  const state: FixtureName = isFixture(data) ? data : 'demo'
  const client = useMemo(() => {
    const seeded = new QueryClient({ defaultOptions: { queries: { staleTime: Infinity, retry: false } } })
    seeded.setQueryData(['procedure', fixtures[state].id], fixtures[state])
    return seeded
  }, [state])

  return (
    <>
      <QueryClientProvider client={client} key={state}>
        <Container size="md" py="md">
          <ProcedurePage />
        </Container>
      </QueryClientProvider>
      <Affix position={{ bottom: 'md', left: '50%' }}>
        <SegmentedControl
          size="xs"
          value={state}
          data={states}
          onChange={(value) => setParams({ data: value }, { replace: true })}
        />
      </Affix>
    </>
  )
}
