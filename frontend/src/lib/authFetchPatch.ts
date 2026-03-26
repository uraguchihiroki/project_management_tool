import { getAuthToken } from '@/lib/authToken'

let installed = false

/**
 * useEffect 後では遅いため、クライアントバンドル読み込み直後に fetch を差し替える。
 * これがないと初回の raw fetch が Authorization なしで飛び 401 になる。
 */
export function installAuthFetchPatch(): void {
  if (typeof window === 'undefined' || installed) return
  installed = true
  const originalFetch = window.fetch.bind(window)
  window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
    const token = getAuthToken()
    if (!token) return originalFetch(input, init)
    const headers = new Headers(init?.headers ?? {})
    if (!headers.has('Authorization')) {
      headers.set('Authorization', `Bearer ${token}`)
    }
    return originalFetch(input, { ...init, headers })
  }
}
