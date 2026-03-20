const TOKEN_KEY = 'authToken'

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
