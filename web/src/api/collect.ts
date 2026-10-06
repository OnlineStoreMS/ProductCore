import client, { unwrap, type PageData } from './client'

export interface ProductCollectTask {
  id: number
  productUrl: string
  platform: string
  platformName: string
  agentJobId: number
  agentId: number
  agentName: string
  productId?: number
  status: string
  message: string
  createdAt: string
}

export async function fetchProductCollects(page = 1, pageSize = 20) {
  const res = await client.get('/product-collects', { params: { page, pageSize } })
  return unwrap<PageData<ProductCollectTask>>(res)
}
