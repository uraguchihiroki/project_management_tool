const TOKEN_KEY = 'authToken'
/** AuthContext とキーを揃える */
const SESSION_USER_KEY = 'currentUser'
const SESSION_ORG_KEY = 'currentOrg'
const SESSION_SUPER_ADMIN_KEY = 'currentSuperAdmin'

let inMemoryToken: string | null = null

export function getAuthToken(): string | null {
  if (inMemoryToken) return inMemoryToken
  if (typeof window === 'undefined') return null
  const stored = sessionStorage.getItem(TOKEN_KEY)
  inMemoryToken = stored
  return stored
}

export function setAuthToken(token: string | null) {
  inMemoryToken = token
  if (typeof window === 'undefined') return
  if (token) {
    sessionStorage.setItem(TOKEN_KEY, token)
  } else {
    sessionStorage.removeItem(TOKEN_KEY)
  }
}

/** ログアウト・API 401 時にトークンと各セッションをまとめて消す */
export function clearAuthSession() {
  inMemoryToken = null
  if (typeof window === 'undefined') return
  sessionStorage.removeItem(TOKEN_KEY)
  sessionStorage.removeItem(SESSION_USER_KEY)
  sessionStorage.removeItem(SESSION_ORG_KEY)
  sessionStorage.removeItem(SESSION_SUPER_ADMIN_KEY)
}
