import { useQuery } from '@tanstack/react-query'
import { api } from '../../api/client'

export function useGreeting() {
  return useQuery({
    queryKey: ['hello'],
    queryFn: async () => {
      const { data, error } = await api.GET('/hello')
      if (error || !data) throw new Error('The greeting could not be loaded')
      return data
    },
  })
}
