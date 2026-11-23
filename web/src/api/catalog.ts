import { queryOptions } from '@tanstack/react-query'
import { api } from './client'

// catalogQuery reads the catalog of procedure types, shared by the pages that
// show it and the one that starts a procedure.
export const catalogQuery = queryOptions({
  queryKey: ['procedure-types'],
  queryFn: async () => {
    const { data, response } = await api.GET('/procedure-types')
    if (data === undefined) {
      throw new Error(`GET /procedure-types answered ${response.status}`)
    }
    return data
  },
})
