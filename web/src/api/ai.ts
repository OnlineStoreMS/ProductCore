import client, { unwrap } from './client'
import type { UploadContext } from './upload'

export interface RetouchImageInput {
  imageUrl: string
  prompt: string
  model?: string
  connectionId?: number
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

export interface AIImageModel {
  id: number
  connectionId: number
  connectionName: string
  providerCode: string
  modelId: string
  displayName: string
  isDefault: number
}

export interface AISetting {
  defaultImageModel: string
  defaultImageConnectionId: number
  defaultImageLabel?: string
}

export async function fetchAIImageModels() {
  const res = await client.get('/ai/image-models')
  return unwrap<AIImageModel[]>(res)
}

export async function fetchAISetting() {
  const res = await client.get('/ai/settings')
  return unwrap<AISetting>(res)
}

export async function updateAISetting(body: { defaultImageModel: string; defaultImageConnectionId?: number }) {
  const res = await client.put('/ai/settings', body)
  return unwrap<AISetting>(res)
}

export function imageModelKey(model: Pick<AIImageModel, 'connectionId' | 'modelId'>) {
  return `${model.connectionId}|${model.modelId}`
}

export function imageModelLabel(model: AIImageModel) {
  const name = model.displayName || model.modelId
  return model.connectionName ? `${model.connectionName} · ${name}` : name
}
