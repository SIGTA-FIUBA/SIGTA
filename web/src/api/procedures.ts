import { useQueries, useQuery } from '@tanstack/react-query'
import { api } from './client'
import { requestedNames } from '../content/procedures'

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

function procedureQuery(id: string) {
  return {
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
  }
}

export function useProcedure(id: string) {
  return useQuery(procedureQuery(id))
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

// useCorrections asks for the detail of each of the person's own procedures
// with a correction pending, the only ones where the applicant has something
// to do: the list says that there is one, the detail what it asks for.
function useCorrections(ids: string[]) {
  return useQueries({ queries: ids.map(procedureQuery) })
}

type Task = { id: string; procedureId: string; task: string; context: string }

// useApplicantTasks is what the person has to do on their own procedures:
// one task for each correction they have not answered, named by what it asks
// for, under the procedure it belongs to.
export function useApplicantTasks() {
  const mine = useMyProcedures()
  const ids = (mine.data ?? []).filter((procedure) => procedure.correction_pending).map((procedure) => procedure.id)
  const details = useCorrections(ids)
  const tasks: Task[] = details.flatMap((detail) =>
    detail.data
      ? detail.data.correction_requests.map((request, index) => ({
          id: `${detail.data.id}-${index}`,
          procedureId: detail.data.id,
          task: requestedNames(detail.data.steps, request) || 'Subsanación pendiente',
          context: detail.data.name,
        }))
      : [],
  )
  return {
    tasks,
    procedures: ids.length,
    isPending: mine.isPending || details.some((detail) => detail.isPending),
    isError: mine.isError || details.some((detail) => detail.isError),
  }
}
