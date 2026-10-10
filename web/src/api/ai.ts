import client, { unwrap } from './client'
import type { UploadContext } from './upload'

export interface RetouchImageInput {
  imageUrl: string
  prompt: string
  scope?: UploadContext['scope']
  resource?: UploadContext['resource']
  productId?: number
  skuId?: number
}

export async function retouchProductImage(input: RetouchImageInput): Promise<{ url: string }> {
  const res = await client.post('/ai/images/retouch', input, { timeout: 180000 })
  return unwrap<{ url: string }>(res)
}
