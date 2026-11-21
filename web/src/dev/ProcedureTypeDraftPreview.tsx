import { Affix, Container, SegmentedControl } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { ProcedureTypeDraftPage } from '../pages/ProcedureTypeDraftPage'
import { fixtures, type FixtureName } from './procedureTypeDraftFixtures'

const fixtureOptions: { value: FixtureName; label: string }[] = [
  { value: 'demo', label: 'Demo' },
  { value: 'proposed', label: 'Proposed' },
  { value: 'worst', label: 'Worst case' },
]

function isFixture(value: string | null): value is FixtureName {
  return value !== null && value in fixtures
}

// ProcedureTypeDraftPreview renders the editor of procedure types with each
// fixture of procedureTypeDraftFixtures, swapped in the query cache, which is
// where the page reads it. Only for development: the route exists only in the
// dev server. Changes still go to the API. The fixture stays in ?data= across
// reloads.
export function ProcedureTypeDraftPreview() {
  const [params, setParams] = useSearchParams()
  const fixtureParam = params.get('data')
  const fixture: FixtureName = isFixture(fixtureParam) ? fixtureParam : 'demo'
  const client = useMemo(() => {
    const data = fixtures[fixture]
    const seeded = new QueryClient({ defaultOptions: { queries: { staleTime: Infinity, retry: false } } })
    seeded.setQueryData(['procedure-types', data.draft.procedure_type_id, 'draft'], data.draft)
    seeded.setQueryData(['editable-procedure-types'], data.editable)
    return seeded
  }, [fixture])

  return (
    <>
      <QueryClientProvider client={client} key={fixture}>
        <Container size="xl" py="md">
          <ProcedureTypeDraftPage />
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
