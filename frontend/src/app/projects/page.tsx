'use client'

import { useQuery } from '@tanstack/react-query'
import { getProjects } from '@/lib/api'
import Link from 'next/link'
import { FolderKanban, ChevronRight } from 'lucide-react'
import type { Project } from '@/types'
import { useRequireAuth, useAuth } from '@/context/AuthContext'
import Header from '@/components/Header'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'

export default function ProjectsPage() {
  const currentUser = useRequireAuth()
  const { currentOrg } = useAuth()
  const authFetch = useAuthFetchEnabled()

  const { data: projects = [], isLoading } = useQuery({
    queryKey: ['projects', currentOrg?.id],
    queryFn: () => getProjects(currentOrg?.id),
    enabled: authFetch && !!currentOrg?.id,
  })

  if (!currentUser) return null

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />

      <main className="max-w-6xl mx-auto px-6 py-8">
        <h2 className="text-2xl font-bold text-gray-900 mb-6">プロジェクト一覧</h2>

        {isLoading ? (
          <div className="text-center py-16 text-gray-500">読み込み中...</div>
        ) : projects.length === 0 ? (
          <div className="text-center py-16 bg-white rounded-xl border border-dashed border-gray-300">
            <FolderKanban className="w-12 h-12 text-gray-300 mx-auto mb-3" />
            <p className="text-gray-500">プロジェクトがありません</p>
            {currentUser.is_admin && (
              <Link
                href="/admin/projects"
                className="mt-4 inline-block text-blue-600 text-sm font-medium hover:underline"
              >
                管理画面でプロジェクトを作成する
              </Link>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {projects.map((project: Project) => (
              <Link
                key={project.id}
                href={`/projects/${project.id}`}
                className="bg-white rounded-xl border border-gray-200 p-6 hover:border-blue-300 hover:shadow-md transition-all group"
              >
                <div className="flex items-start justify-between">
                  <div>
                    <span className="text-xs font-mono font-bold text-blue-600 bg-blue-50 px-2 py-0.5 rounded">
                      {project.key}
                    </span>
                    <h3 className="mt-2 text-lg font-semibold text-gray-900 group-hover:text-blue-600">
                      {project.name}
                    </h3>
                    {project.description && (
                      <p className="mt-1 text-sm text-gray-500 line-clamp-2">{project.description}</p>
                    )}
                  </div>
                  <ChevronRight className="w-5 h-5 text-gray-400 group-hover:text-blue-500 flex-shrink-0" />
                </div>
                <div className="mt-4 text-xs text-gray-400">
                  オーナー: {project.owner?.name}
                </div>
              </Link>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
