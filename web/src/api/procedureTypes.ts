import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import { change, forbiddenOr } from './requests'
import type { components } from './schema'

type ProcedureTypeDefinition = components['schemas']['ProcedureTypeDefinition']

// Each key follows the route of the API. A change refreshes the list and the
// draft, since both show its status.
const editableKey = ['editable-procedure-types']
const draftKey = (procedureTypeId: number) => ['procedure-types', procedureTypeId, 'draft']

// useEditableProcedureTypes asks what the person can edit or approve. The
// menu shares it to know whether to offer the page, so it costs no request.
export function useEditableProcedureTypes(enabled = true) {
  return useQuery({
    queryKey: editableKey,
    enabled,
    queryFn: async () => {
      const { data, response } = await api.GET('/editable-procedure-types')
      return forbiddenOr('/editable-procedure-types', data, response)
    },
  })
}

export function useProcedureTypeDraft(procedureTypeId: number) {
  return useQuery({
    queryKey: draftKey(procedureTypeId),
    queryFn: async () => {
      const path = { procedureTypeId }
      const { data, response } = await api.GET('/procedure-types/{procedureTypeId}/draft', { params: { path } })
      return forbiddenOr(`/procedure-types/${procedureTypeId}/draft`, data, response)
    },
  })
}

// The changes to a draft. Each throws ActionError when the API refuses it, and
// each but the creation names the revision it was made on.
export const draftActions = {
  create: (definition: ProcedureTypeDefinition) => change(api.POST('/procedure-types', { body: { definition } })),
  open: (procedureTypeId: number) =>
    change(api.POST('/procedure-types/{procedureTypeId}/draft', { params: { path: { procedureTypeId } } })),
  edit: (procedureTypeId: number, revision: number, definition: ProcedureTypeDefinition) =>
    change(
      api.PUT('/procedure-types/{procedureTypeId}/draft', {
        params: { path: { procedureTypeId } },
        body: { revision, definition },
      }),
    ),
  discard: (procedureTypeId: number, revision: number) =>
    change(
      api.DELETE('/procedure-types/{procedureTypeId}/draft', {
        params: { path: { procedureTypeId }, query: { revision } },
      }),
    ),
  propose: (procedureTypeId: number, revision: number) =>
    change(
      api.POST('/procedure-types/{procedureTypeId}/draft/proposal', {
        params: { path: { procedureTypeId } },
        body: { revision },
      }),
    ),
  reject: (procedureTypeId: number, revision: number, reason: string) =>
    change(
      api.POST('/procedure-types/{procedureTypeId}/draft/rejection', {
        params: { path: { procedureTypeId } },
        body: { revision, reason },
      }),
    ),
  approve: (procedureTypeId: number, revision: number) =>
    change(
      api.POST('/procedure-types/{procedureTypeId}/draft/approval', {
        params: { path: { procedureTypeId } },
        body: { revision },
      }),
    ),
}

// useDraftChange runs one of the actions and then refreshes what the pages of
// procedure types show.
export function useDraftChange<T, R>(action: (input: T) => Promise<R>) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: action,
    onSettled: () =>
      queryClient
        .invalidateQueries({ queryKey: ['procedure-types'] })
        .then(() => queryClient.invalidateQueries({ queryKey: editableKey })),
  })
}
