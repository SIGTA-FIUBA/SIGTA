import { queryOptions } from '@tanstack/react-query'
import { api } from './client'

// catalogQuery reads the catalog of procedure types, shared by the pages that
// show it and the one that starts a procedure.
export const catalogQuery = queryOptions({
  queryKey: ['procedure-types'],
  queryFn: async () => {
    const { data, error } = await api.GET('/procedure-types')
    if (error !== undefined || data === undefined) {
      throw new Error('The catalog could not be loaded')
    }
    return data
  },
})
