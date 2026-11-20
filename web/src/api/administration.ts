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

// Each key follows the route of the API. A change refreshes every query of the
// administration, since ending a membership can free an area to be deleted:
// the memberships of an area share the prefix of the areas.
const areasKey = ['areas']
const administratorsKey = ['administrators']

export function useAreas() {
  return useQuery({
    queryKey: areasKey,
    queryFn: async () => {
      const { data, response } = await api.GET('/areas')
      return forbiddenOr('/areas', data, response)
    },
  })
}

export function useAreaMemberships(areaId: string) {
  return useQuery({
    queryKey: [...areasKey, areaId, 'memberships'],
    queryFn: async () => {
      const { data, response } = await api.GET('/areas/{areaId}/memberships', { params: { path: { areaId } } })
      return forbiddenOr(`/areas/${areaId}/memberships`, data, response)
    },
  })
}

export function useAdministrators() {
  return useQuery({
    queryKey: administratorsKey,
    queryFn: async () => {
      const { data, response } = await api.GET('/administrators')
      return forbiddenOr('/administrators', data, response)
    },
  })
}

// searchesPeople tells whether the server searches people with text: it
// needs 2 characters, counted as the server counts them, so an emoji is one.
export function searchesPeople(text: string) {
  return [...text.trim()].length >= 2
}

// usePeopleSearch looks for who entered SIGTA by part of the name or email.
// The server folds case and accents and answers up to 10 people, so the list
// is not filtered again here.
export function usePeopleSearch(text: string) {
  const query = text.trim()
  return useQuery({
    queryKey: ['people', query],
    enabled: searchesPeople(query),
    queryFn: async () => {
      const { data, response } = await api.GET('/people', { params: { query: { q: query } } })
      return forbiddenOr('/people', data, response)
    },
  })
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
  changeRole: (areaId: string, roleId: string, body: { permissions: string[] }) =>
    run(api.PUT('/areas/{areaId}/roles/{roleId}/permissions', { params: { path: { areaId, roleId } }, body })),
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
    onSettled: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: areasKey }),
        queryClient.invalidateQueries({ queryKey: administratorsKey }),
      ]),
  })
}
