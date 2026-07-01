import axios from 'axios'
import type { AxiosInstance } from 'axios'
import { getToken, redirectToPortal, clearToken } from '../utils/auth'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

export interface PageData<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

const client: AxiosInstance = axios.create({
  baseURL: '/api/v1/admin',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

client.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

client.interceptors.response.use(
  (res) => {
    if (res.config.responseType === 'blob') {
      return res
    }
    const body = res.data as ApiResponse
    if (body.code !== 200) {
      return Promise.reject(new Error(body.message || '请求失败'))
    }
    return res
  },
  (err) => {
    if (err.response?.status === 401) {
      clearToken()
      redirectToPortal()
    }
    return Promise.reject(err)
  },
)

export function unwrap<T>(res: { data: ApiResponse<T> }): T {
  return res.data.data as T
}

export default client

export function parseContentDispositionFilename(header?: string): string | undefined {
  if (!header) return undefined
  const star = header.match(/filename\*=UTF-8''([^;\n]+)/i)
  if (star?.[1]) {
    try {
      return decodeURIComponent(star[1])
    } catch {
      /* ignore */
    }
  }
  const quoted = header.match(/filename="([^"]+)"/)
  if (quoted?.[1]) return quoted[1]
  return undefined
}
