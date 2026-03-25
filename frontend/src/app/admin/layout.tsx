'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Settings, Users, GitBranch, FileText, FolderKanban, Building2 } from 'lucide-react'
import Header from '@/components/Header'
import { useRequireAdmin } from '@/context/AuthContext'

const navItems = [
  { href: '/admin/projects', label: 'プロジェクト管理', icon: FolderKanban },
  { href: '/admin/groups', label: 'グループ管理', icon: Building2 },
  { href: '/admin/users', label: 'ユーザー管理', icon: Users },
  { href: '/admin/workflows', label: 'ワークフロー', icon: GitBranch },
  { href: '/admin/templates', label: 'Issueテンプレート', icon: FileText },
]

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const currentUser = useRequireAdmin()
  const pathname = usePathname()

  // currentUser は AuthContext の useLayoutEffect で sessionStorage から復元されるまで null。
  // ここで return null すると初回ペイントが真っ白になり「何も表示されない」ように見える。
  if (!currentUser) {
    return (
      <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center gap-2 text-gray-600">
        <span className="text-sm">読み込み中…</span>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header
        title={
          <div className="flex items-center gap-2 text-sm text-gray-600">
            <Settings className="w-4 h-4" />
            <span className="font-medium">管理画面</span>
          </div>
        }
        backHref="/projects"
      />

      <div className="flex">
        {/* サイドナビ */}
        <aside className="w-52 min-h-[calc(100vh-65px)] bg-white border-r border-gray-200 pt-4 flex-shrink-0">
          <nav className="px-3 space-y-1">
            {navItems.map(({ href, label, icon: Icon }) => (
              <Link
                key={href}
                href={href}
                className={`flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors ${
                  pathname === href ||
                  (href === '/admin/projects' && pathname.startsWith('/admin/projects/')) ||
                  (href === '/admin/workflows' && pathname.startsWith('/admin/workflows'))
                    ? 'bg-blue-50 text-blue-700 font-medium'
                    : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                }`}
              >
                <Icon className="w-4 h-4 flex-shrink-0" />
                {label}
              </Link>
            ))}
          </nav>
        </aside>

        {/* コンテンツ */}
        <main className="flex-1 p-8 min-w-0">{children}</main>
      </div>
    </div>
  )
}
