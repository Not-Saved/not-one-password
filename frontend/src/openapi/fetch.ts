import createClient from 'openapi-fetch'
import type { paths } from './openapi'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL

export const clientWithoutRefresh = createClient<paths>({
  baseUrl: API_BASE_URL,
  credentials: 'include',
})

export const client = createClient<paths>({
  baseUrl: API_BASE_URL,
  fetch: async (input) => {
    const response = await fetch(input)
    if (response.status === 401) {
      const refreshResponse = await clientWithoutRefresh.POST('/refresh')
      if (refreshResponse.response.ok) {
        const retryResponse = await fetch(input)
        return retryResponse
      }
    }
    return response
  },
  credentials: 'include',
})
