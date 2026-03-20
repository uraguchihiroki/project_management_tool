'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState } from 'react'
import { AuthProvider } from '@/context/AuthContext'
import { installAuthFetchPatch } from '@/lib/authFetchPatch'

installAuthFetchPatch()

export default function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          // localhost 開発で navigator.onLine が false になると
          // mutation が pause して axios が一切飛ばないことがある（バックエンドにログも出ない）
          queries: { staleTime: 1000 * 60, networkMode: 'always' },
          mutations: { networkMode: 'always' },
        },
      })
  )

  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        {children}
      </AuthProvider>
    </QueryClientProvider>
  )
}
