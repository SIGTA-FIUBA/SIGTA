import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import { ForbiddenError } from './procedures'
import type { components } from './schema'

type NewMembership = components['schemas']['NewMembership']
type NewRole = components['schemas']['NewRole']

// ActionError carries the API's message, which says why a change was refused
// and what to do.
class ActionError extends Error {}

const connectionFailure = 'No pudimos guardar el cambio. Revisá tu conexión y volvé a intentarlo.'
const unknownFailure = 'No pudimos guardar el cambio. Volvé a intentarlo.'

async function run(call: Promise<{ error?: unknown; response: Response }>) {
  let outcome: { error?: unknown; response: Response }
  try {
    outcome = await call
  } catch {
    throw new ActionError(connectionFailure)
  }
  if (!outcome.response.ok) {
    const { error } = outcome
    const message =
      typeof error === 'object' && error !== null && 'message' in error && typeof error.message === 'string'
        ? error.message
        : unknownFailure
    throw new ActionError(message)
  }
}

function forbiddenOr<T>(path: string, data: T | undefined, response: Response): T {
  if (response.status === 403) {
    throw new ForbiddenError()
  }
  if (data === undefined) {
    throw new Error(`GET ${path} answered ${response.status}`)
  }
  return data
}

// Every query of the administration shares this prefix, so a change refreshes
// all of them: ending a membership can free an area to be deleted.
const administration = ['administration']

export function useAreas() {
  return useQuery({
    queryKey: [...administration, 'areas'],
    queryFn: async () => {
      const { data, response } = await api.GET('/areas')
      return forbiddenOr('/areas', data, response)
    },
  })
}

export function useAreaMemberships(areaId: string) {
  return useQuery({
    queryKey: [...administration, 'areas', areaId, 'memberships'],
    queryFn: async () => {
      const { data, response } = await api.GET('/areas/{areaId}/memberships', { params: { path: { areaId } } })
      return forbiddenOr(`/areas/${areaId}/memberships`, data, response)
    },
  })
}

export function useAdministrators() {
  return useQuery({
    queryKey: [...administration, 'administrators'],
    queryFn: async () => {
      const { data, response } = await api.GET('/administrators')
      return forbiddenOr('/administrators', data, response)
    },
  })
}

// findPerson looks up someone who has entered SIGTA, by the whole email. It
// returns null when nobody registered with it.
export async function findPerson(email: string) {
  let outcome
  try {
    outcome = await api.GET('/people', { params: { query: { email } } })
  } catch {
    throw new ActionError('No pudimos buscar a la persona. Revisá tu conexión y volvé a intentarlo.')
  }
  if (outcome.response.status === 404) {
    return null
  }
  if (outcome.data === undefined) {
    throw new ActionError('No pudimos buscar a la persona. Volvé a intentarlo.')
  }
  return outcome.data
}

// The changes an administrator can make. Each throws ActionError when the API
// refuses it.
export const actions = {
  createArea: (body: { id: string; name: string }) => run(api.POST('/areas', { body })),
  deleteArea: (areaId: string) => run(api.DELETE('/areas/{areaId}', { params: { path: { areaId } } })),
  createPermission: (areaId: string, body: { id: string; name: string }) =>
    run(api.POST('/areas/{areaId}/permissions', { params: { path: { areaId } }, body })),
  deletePermission: (areaId: string, permissionId: string) =>
    run(api.DELETE('/areas/{areaId}/permissions/{permissionId}', { params: { path: { areaId, permissionId } } })),
  createRole: (areaId: string, body: NewRole) =>
    run(api.POST('/areas/{areaId}/roles', { params: { path: { areaId } }, body })),
  deleteRole: (areaId: string, roleId: string) =>
    run(api.DELETE('/areas/{areaId}/roles/{roleId}', { params: { path: { areaId, roleId } } })),
  grantMembership: (body: NewMembership) => run(api.POST('/memberships', { body })),
  changeMembership: (membershipId: string, body: { roles: string[]; permissions: string[] }) =>
    run(api.PUT('/memberships/{membershipId}/permissions', { params: { path: { membershipId } }, body })),
  endMembership: (membershipId: string) =>
    run(api.DELETE('/memberships/{membershipId}', { params: { path: { membershipId } } })),
  grantAdministrator: (userId: string) => run(api.POST('/administrators', { body: { user_id: userId } })),
  endAdministrator: (administratorId: string) =>
    run(api.DELETE('/administrators/{administratorId}', { params: { path: { administratorId } } })),
}

// useAdministrationChange runs one of the actions and then refreshes what the
// administration shows.
export function useAdministrationChange<T>(change: (input: T) => Promise<void>) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: change,
    onSettled: () => queryClient.invalidateQueries({ queryKey: administration }),
  })
}
