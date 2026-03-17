'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthContext'
import { getUserOrganizations } from '@/lib/api'
import { Building2, ChevronRight } from 'lucide-react'
import type { Organization } from '@/types'

export default function SelectOrgPage() {
  const { currentUser, selectOrg } = useAuth()
  const router = useRouter()
  const [orgs, setOrgs] = useState<Organization[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!currentUser) {
      router.push('/login')
      return
    }
    getUserOrganizations(currentUser.id)
      .then((data) => {
        setOrgs(data)
        if (data.length === 1) {
          selectOrg(data[0])
          router.push('/projects')
        }
      })
      .finally(() => setLoading(false))
  }, [currentUser, router, selectOrg])

  const handleSelect = (org: Organization) => {
    selectOrg(org)
    router.push('/projects')
  }

  if (loading) {
    return <div className="flex items-center justify-center h-screen text-gray-500">読み込み中...</div>
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="flex items-center justify-center gap-2 mb-8">
          <Building2 className="w-8 h-8 text-blue-600" />
          <span className="text-2xl font-bold text-gray-900">組織を選択</span>
        </div>

        <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100">
            <p className="text-sm text-gray-500">
              ログインする組織を選択してください（{currentUser?.name}）
            </p>
          </div>

          {orgs.length === 0 ? (
            <div className="p-8 text-center text-gray-400">
              <Building2 className="w-10 h-10 mx-auto mb-3 text-gray-300" />
              <p className="text-sm">所属している組織がありません</p>
              <p className="text-xs mt-1">管理者に組織への追加を依頼してください</p>
            </div>
          ) : (
            <ul className="divide-y divide-gray-100">
              {orgs.map((org) => (
                <li key={org.id}>
                  <button
                    onClick={() => handleSelect(org)}
                    className="w-full flex items-center justify-between px-6 py-4 hover:bg-blue-50 transition-colors group text-left"
                  >
                    <div className="flex items-center gap-3">
                      <div className="w-9 h-9 rounded-lg bg-blue-100 flex items-center justify-center flex-shrink-0">
                        <Building2 className="w-5 h-5 text-blue-600" />
                      </div>
                      <span className="font-medium text-gray-900 group-hover:text-blue-700">
                        {org.name}
                      </span>
                    </div>
                    <ChevronRight className="w-4 h-4 text-gray-400 group-hover:text-blue-600" />
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  )
}
