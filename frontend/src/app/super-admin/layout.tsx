'use client'

import { SuperAdminProvider } from '@/context/SuperAdminContext'

export default function SuperAdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <SuperAdminProvider>
      {children}
    </SuperAdminProvider>
  )
}
