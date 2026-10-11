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

export interface EraseRegion {
  box: number[]
  polygon?: number[][] | null
  contours?: number[][][] | null
}

export interface EraseImageInput {
  imageUrl: string
  boxes: number[][]
  polygons: Array<number[][] | null>
  shapes: string[]
  scope?: UploadContext['scope']
  resource?: UploadContext['resource']
  productId?: number
  skuId?: number
}

export async function snapProductImage(input: Pick<EraseImageInput, 'imageUrl' | 'boxes' | 'polygons'>): Promise<{ regions: EraseRegion[] }> {
  const res = await client.post('/ai/images/erase/snap', input, { timeout: 60000 })
  return unwrap<{ regions: EraseRegion[] }>(res)
}

export async function eraseProductImage(input: EraseImageInput): Promise<{ url: string }> {
  const res = await client.post('/ai/images/erase', input, { timeout: 60000 })
  return unwrap<{ url: string }>(res)
}
