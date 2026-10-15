import axios from 'axios'
import { getToken, tryRefreshAccessToken, clearToken, redirectToPortal } from '../utils/auth'
import type { ApiResponse, PageData } from './client'

export interface MaterialCategory {
  id: number
  parentId: number
  name: string
  code: string
  children?: MaterialCategory[]
}

export interface MaterialItem {
  id: number
  categoryId: number
  title: string
  mediaType: 'image' | 'video'
  url: string
  coverUrl: string
  thumbUrl?: string
  fileName: string
  width: number
  height: number
  sizeBytes: number
}

function materialBase(): string {
  return '/apps/material/api/v1/admin'
}

const materialClient = axios.create({
  baseURL: materialBase(),
  timeout: 60000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
})

materialClient.interceptors.request.use((config) => {
  const token = getToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

materialClient.interceptors.response.use(
  (res) => {
    const body = res.data as ApiResponse
    if (body.code !== 200) {
      return Promise.reject(new Error(body.message || '请求失败'))
    }
    return res
  },
  async (err) => {
    const cfg = err.config as { _retry?: boolean } | undefined
    if (err.response?.status === 401 && cfg && !cfg._retry) {
      const ok = await tryRefreshAccessToken()
      if (ok) {
        cfg._retry = true
        return materialClient.request(cfg as any)
      }
      clearToken()
      redirectToPortal()
    }
    const msg = err.response?.data?.message || err.message || '请求失败'
    return Promise.reject(new Error(msg))
  },
)

function unwrap<T>(res: { data: ApiResponse<T> }): T {
  return res.data.data as T
}

export async function listMaterialCategories(): Promise<MaterialCategory[]> {
  const res = await materialClient.get('/categories')
  return unwrap(res) || []
}

export async function listMaterials(params?: Record<string, unknown>): Promise<PageData<MaterialItem>> {
  const res = await materialClient.get('/materials', { params })
  return unwrap(res)
}

/** 素材文件 URL 补全（相对路径） */
export function materialAssetUrl(url?: string) {
  if (!url) return ''
  if (/^https?:\/\//i.test(url) || url.startsWith('data:')) return url
  if (url.startsWith('/apps/')) return url
  if (url.startsWith('/')) return `/apps/material${url}`
  return `/apps/material/${url}`
}

export function materialListSrc(m: Pick<MaterialItem, 'thumbUrl' | 'coverUrl' | 'url' | 'mediaType'>) {
  if (m.mediaType === 'video') return materialAssetUrl(m.coverUrl || '')
  return materialAssetUrl(m.thumbUrl || m.coverUrl || m.url || '')
}
