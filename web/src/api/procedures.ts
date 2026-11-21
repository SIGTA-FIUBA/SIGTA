import { useQuery } from '@tanstack/react-query'
import { api } from './client'

// ForbiddenError means the person is not part of the staff.
export class ForbiddenError extends Error {}

// useInboxProcedures asks once for every procedure in progress the person can see.
// The inbox and the menu share it, so the pending count costs no request.
export function useInboxProcedures(enabled = true) {
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

// NotFoundError means the procedure does not exist or the person cannot see it:
// the API answers both the same.
export class NotFoundError extends Error {}

export function useProcedure(id: string) {
  return useQuery({
    queryKey: ['procedure', id],
    queryFn: async () => {
      const { data, response } = await api.GET('/procedures/{procedureId}', { params: { path: { procedureId: id } } })
      if (response.status === 404) {
        throw new NotFoundError()
      }
      if (data === undefined) {
        throw new Error(`GET /procedures/${id} answered ${response.status}`)
      }
      return data
    },
  })
}

// useMyProcedures asks for the procedures that the person started, staff or
// not, in progress and closed.
export function useMyProcedures() {
  return useQuery({
    queryKey: ['me', 'procedures'],
    queryFn: async () => {
      const { data, response } = await api.GET('/me/procedures')
      if (data === undefined) {
        throw new Error(`GET /me/procedures answered ${response.status}`)
      }
      return data
    },
  })
}
