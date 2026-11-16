import { Affix, Container, SegmentedControl } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { catalogQuery } from '../api/catalog'
import { StartProcedurePage } from '../pages/StartProcedurePage'
import { fixtures, type FixtureName } from './startProcedureFixtures'

const states: { value: FixtureName; label: string }[] = [
  { value: 'demo', label: 'Demo' },
  { value: 'worst', label: 'Worst case' },
  { value: 'empty', label: 'Empty' },
  { value: 'one', label: 'One' },
]

function isFixture(value: string | null): value is FixtureName {
  return value !== null && value in fixtures
}

// StartProcedurePreview renders the form that starts a procedure with each
// catalog of startProcedureFixtures, swapped in the query cache, which is
// where the page reads the catalog. Only for development: the route exists
// only in the dev server. The state stays in ?data= across reloads.
export function StartProcedurePreview() {
  const [params, setParams] = useSearchParams()
  const data = params.get('data')
  const state: FixtureName = isFixture(data) ? data : 'demo'
  const client = useMemo(() => {
    const seeded = new QueryClient({ defaultOptions: { queries: { staleTime: Infinity, retry: false } } })
    seeded.setQueryData(catalogQuery.queryKey, [fixtures[state]])
    return seeded
  }, [state])

  return (
    <>
      <QueryClientProvider client={client} key={state}>
        <Container size="md" py="md">
          <StartProcedurePage />
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
