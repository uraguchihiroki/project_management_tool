'use client'

import { useState, useEffect, useRef } from 'react'
import Link from 'next/link'
import { FolderKanban, User, LogOut, ArrowLeft, Settings, Building2, ChevronDown, Check } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'
import { getUserOrganizations } from '@/lib/api'
import type { Organization } from '@/types'

interface HeaderProps {
  backHref?: string
  title?: React.ReactNode
  actions?: React.ReactNode
}

export default function Header({ backHref, title, actions }: HeaderProps) {
  const { currentUser, currentOrg, logout, selectOrg } = useAuth()
  const [orgs, setOrgs] = useState<Organization[]>([])
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // ユーザーの所属組織を取得
  useEffect(() => {
    if (currentUser) {
      getUserOrganizations(currentUser.id)
        .then(setOrgs)
        .catch(() => {})
    }
  }, [currentUser])

  // ドロップダウン外クリックで閉じる
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const handleOrgSelect = async (org: Organization) => {
    await selectOrg(org)
    setDropdownOpen(false)
    window.location.reload()
  }

  const canSwitchOrg = orgs.length > 1

  return (
    <header className="bg-white border-b border-gray-200 px-6 py-4 sticky top-0 z-10">
      <div className="max-w-full mx-auto flex items-center justify-between gap-4">
        {/* 左側: ロゴ / 戻るボタン + タイトル */}
        <div className="flex items-center gap-4 min-w-0">
          {backHref ? (
            <Link href={backHref} className="text-gray-400 hover:text-gray-600 flex-shrink-0">
              <ArrowLeft className="w-5 h-5" />
            </Link>
          ) : (
            <Link href="/projects" className="flex items-center gap-2 flex-shrink-0">
              <FolderKanban className="w-5 h-5 text-blue-600" />
              <span className="text-base font-bold text-gray-900 hidden sm:block">ProjectHub</span>
            </Link>
          )}
          {title && <div className="min-w-0">{title}</div>}
        </div>

        {/* 右側: 追加アクション + 組織 + ユーザー情報 */}
        <div className="flex items-center gap-3 flex-shrink-0">
          {actions}
          {currentUser && (
            <>
              {/* 組織表示 */}
              {currentOrg && (
                <div className="relative" ref={dropdownRef}>
                  <button
                    onClick={() => canSwitchOrg && setDropdownOpen(!dropdownOpen)}
                    className={`flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg border transition-colors
                      ${canSwitchOrg
                        ? 'border-blue-200 bg-blue-50 text-blue-700 hover:bg-blue-100 cursor-pointer'
                        : 'border-gray-200 bg-gray-50 text-gray-600 cursor-default'
                      }`}
                  >
                    <Building2 className="w-4 h-4 flex-shrink-0" />
                    <span className="font-medium hidden sm:block max-w-32 truncate">
                      {currentOrg.name}
                    </span>
                    {canSwitchOrg && <ChevronDown className="w-3 h-3" />}
                  </button>

                  {/* 組織切替ドロップダウン */}
                  {dropdownOpen && (
                    <div className="absolute right-0 mt-1 w-56 bg-white rounded-xl border border-gray-200 shadow-lg overflow-hidden z-50">
                      <div className="px-3 py-2 border-b border-gray-100">
                        <p className="text-xs text-gray-400 font-medium">組織を切り替える</p>
                      </div>
                      <ul className="py-1">
                        {orgs.map((org) => (
                          <li key={org.id}>
                            <button
                              onClick={() => handleOrgSelect(org)}
                              className="w-full flex items-center gap-2 px-3 py-2 hover:bg-blue-50 transition-colors text-left"
                            >
                              <Building2 className="w-4 h-4 text-gray-400 flex-shrink-0" />
                              <span className="flex-1 text-sm text-gray-700 truncate">{org.name}</span>
                              {currentOrg.id === org.id && (
                                <Check className="w-4 h-4 text-blue-600 flex-shrink-0" />
                              )}
                            </button>
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              )}

              {/* 管理リンク */}
              {currentUser.is_admin && (
                <Link
                  href="/admin"
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-500 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                  title="管理画面"
                >
                  <Settings className="w-4 h-4" />
                  <span className="hidden sm:block">管理</span>
                </Link>
              )}

              {/* ユーザー情報 */}
              <div className="flex items-center gap-2 text-sm text-gray-600 bg-gray-100 px-3 py-1.5 rounded-lg">
                <User className="w-4 h-4" />
                <span className="font-medium">{currentUser.name}</span>
                {currentUser.is_admin && (
                  <span className="text-xs bg-blue-100 text-blue-700 px-1.5 py-0.5 rounded font-medium">管理者</span>
                )}
              </div>

              {/* ログアウト */}
              <button
                onClick={logout}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                title="ログアウト"
              >
                <LogOut className="w-4 h-4" />
                <span className="hidden sm:block">ログアウト</span>
              </button>
            </>
          )}
        </div>
      </div>
    </header>
  )
}
