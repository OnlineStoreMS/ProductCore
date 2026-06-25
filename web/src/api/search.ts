import client, { unwrap, type PageData } from './client'
import type { SuperSearchItem } from '../types/search'

export interface SuperSearchQuery {
  keyword: string
  page?: number
  pageSize?: number
}

export async function superSearch(query: SuperSearchQuery) {
  const res = await client.get('/super-search', { params: query })
  return unwrap<PageData<SuperSearchItem>>(res)
}
