import { useQuery } from '@tanstack/react-query'
import { api } from './client'

// ForbiddenError means the person is not part of the staff.
export class ForbiddenError extends Error {}

// useProcedures asks once for every procedure in progress the person can see.
// The inbox and the menu share it, so the pending count costs no request.
export function useProcedures(enabled = true) {
  return useQuery({
    queryKey: ['procedures'],
    enabled,
    queryFn: async () => {
      const { data, response } = await api.GET('/procedures', { params: { query: { view: 'all' } } })
      if (response.status === 403) {
        throw new ForbiddenError()
      }
      if (data === undefined) {
        throw new Error(`GET /procedures answered ${response.status}`)
      }
      return data
    },
  })
}
