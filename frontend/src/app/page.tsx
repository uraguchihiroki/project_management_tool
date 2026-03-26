import Link from 'next/link'
import { LayoutDashboard } from 'lucide-react'

export default function HomePage() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="text-center space-y-6">
        <div className="flex items-center justify-center gap-3">
          <LayoutDashboard className="w-12 h-12 text-blue-600" />
          <h1 className="text-4xl font-bold text-gray-900">ProjectHub</h1>
        </div>
        <p className="text-lg text-gray-600">チケットベースのプロジェクト管理ツール</p>
        <Link
          href="/projects"
          className="inline-block px-8 py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors"
        >
          プロジェクト一覧へ
        </Link>
      </div>
    </div>
  )
}
