const TOKEN_KEY = 'uc_access_token'
const PORTAL_URL = import.meta.env.VITE_PORTAL_URL || 'http://localhost:5174'

let sessionVerified = false

export interface SessionUser {
  id: number
  email: string
  displayName: string
  isPlatform: boolean
}

export interface SessionTenant {
  id: number
  companyId: number
  name: string
  code: string
}

export interface SessionInfo {
  user: SessionUser
  tenant: SessionTenant
  tenants: SessionTenant[]
}

export function getToken(): string | undefined {
  return localStorage.getItem(TOKEN_KEY) || undefined
}

export function saveToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
  sessionVerified = false
  sessionStorage.removeItem('uc_session_profile')
}

export function resetSessionVerification() {
  sessionVerified = false
}

/** 刚从 UserCore 跳转带 token 进入，跳过首次 /auth/me 校验 */
export function trustFreshToken() {
  sessionVerified = true
}

export function redirectToPortal() {
  window.location.href = `${PORTAL_URL}/login`
}

export function portalAppsUrl() {
  return `${PORTAL_URL}/apps`
}

export function portalLoginUrl() {
  return `${PORTAL_URL}/login`
}

/** 向 UserCore 校验 JWT 是否仍有效（过期/篡改会失败） */
export async function verifySession(): Promise<boolean> {
  const { fetchSession } = await import('../api/session')
  const info = await fetchSession()
  return info !== null
}

/** 路由守卫：有 token 且校验通过 */
export async function ensureSession(): Promise<boolean> {
  if (!getToken()) return false
  if (sessionVerified) return true
  const ok = await verifySession()
  if (ok) sessionVerified = true
  return ok
}
