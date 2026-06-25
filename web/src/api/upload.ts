import client, { unwrap } from './client'

export interface BatchUploadFailed {
  filename: string
  message: string
}

export interface BatchUploadResult {
  urls: string[]
  failed: BatchUploadFailed[]
}

/** 上传资源类型，对应后端 storage 路径层级 */
export type UploadResource =
  | 'main'
  | 'album'
  | 'pics34'
  | 'detail'
  | 'material_white'
  | 'material_transparent'
  | 'material_guide34'
  | 'material_long'
  | 'video_11'
  | 'video_34'
  | 'video_169'
  | 'video_916'
  | 'spec'
  | 'platform_logo'

export interface UploadContext {
  scope?: 'product' | 'common'
  productId?: number
  skuId?: number
  resource?: UploadResource | string
}

function appendUploadContext(form: FormData, ctx?: UploadContext) {
  if (!ctx) return
  if (ctx.scope) form.append('scope', ctx.scope)
  if (ctx.resource) form.append('resource', ctx.resource)
  if (ctx.productId) form.append('productId', String(ctx.productId))
  if (ctx.skuId) form.append('skuId', String(ctx.skuId))
}

export async function uploadImage(file: File, ctx?: UploadContext): Promise<string> {
  const form = new FormData()
  form.append('file', file)
  appendUploadContext(form, ctx)
  const res = await client.post('/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  const data = unwrap<{ url: string }>(res)
  return data.url
}

export async function uploadImagesBatch(files: File[], ctx?: UploadContext): Promise<BatchUploadResult> {
  const form = new FormData()
  for (const file of files) {
    form.append('files', file)
  }
  appendUploadContext(form, ctx)
  const res = await client.post('/upload/batch', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 600000,
  })
  return unwrap<BatchUploadResult>(res)
}

export async function uploadImageFromUrl(sourceUrl: string, ctx?: UploadContext): Promise<string> {
  const res = await client.post('/upload/from-url', {
    url: sourceUrl,
    scope: ctx?.scope,
    resource: ctx?.resource,
    productId: ctx?.productId,
    skuId: ctx?.skuId,
  })
  const data = unwrap<{ url: string }>(res)
  return data.url
}

export async function uploadVideo(file: File, ctx?: UploadContext): Promise<string> {
  const form = new FormData()
  form.append('file', file)
  appendUploadContext(form, ctx)
  const res = await client.post('/upload/video', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 600000,
  })
  const data = unwrap<{ url: string }>(res)
  return data.url
}
