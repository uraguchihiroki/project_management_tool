'use client'

import { useState, useEffect } from 'react'
import { getAuthToken } from '@/lib/authToken'

/**
 * Next.js の SSR ではトークンが付かず 401 になるのを防ぐ。
 * マウント後かつ sessionStorage の JWT があるときだけ API を叩く。
 */
export function useAuthFetchEnabled(): boolean {
  const [mounted, setMounted] = useState(false)
  useEffect(() => {
    setMounted(true)
  }, [])
  return mounted && !!getAuthToken()
}
