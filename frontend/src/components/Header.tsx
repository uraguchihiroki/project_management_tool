'use client'

import Link from 'next/link'
import { FolderKanban, User, LogOut, ArrowLeft } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'

interface HeaderProps {
  /** 戻るボタンのリンク先。指定しない場合は表示しない */
  backHref?: string
  /** ページタイトル（パンくずなど） */
  title?: React.ReactNode
  /** ヘッダー右側に置く追加ボタンなど */
  actions?: React.ReactNode
}

export default function Header({ backHref, title, actions }: HeaderProps) {
  const { currentUser, logout } = useAuth()

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

        {/* 右側: 追加アクション + ユーザー情報 */}
        <div className="flex items-center gap-3 flex-shrink-0">
          {actions}
          {currentUser && (
            <>
              <div className="flex items-center gap-2 text-sm text-gray-600 bg-gray-100 px-3 py-1.5 rounded-lg">
                <User className="w-4 h-4" />
                <span className="font-medium">{currentUser.name}</span>
              </div>
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
