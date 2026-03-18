'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getProjects, createProject } from '@/lib/api'
import { useState } from 'react'
import Link from 'next/link'
import { Plus, FolderKanban, ChevronRight } from 'lucide-react'
import type { Project, ProjectStatus } from '@/types'

const PROJECT_STATUS_LABELS: Record<ProjectStatus, string> = {
  none: 'なし',
  planning: '計画中',
  active: '実行中',
  completed: '完了',
}
import { useRequireAdmin, useAuth } from '@/context/AuthContext'

export default function AdminProjectsPage() {
  const currentUser = useRequireAdmin()
  const { currentOrg } = useAuth()
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({
    key: '',
    name: '',
    description: '',
    start_date: '',
    end_date: '',
    status: 'none' as ProjectStatus,
  })

  const { data: projects = [], isLoading } = useQuery({
    queryKey: ['projects', currentOrg?.id],
    queryFn: () => getProjects(currentOrg?.id),
  })

  const createMutation = useMutation({
    mutationFn: (data: Parameters<typeof createProject>[0]) => createProject(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      setShowForm(false)
      setForm({ key: '', name: '', description: '', start_date: '', end_date: '', status: 'none' })
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!currentUser) return
    createMutation.mutate({
      key: form.key.toUpperCase(),
      name: form.name,
      description: form.description || undefined,
      owner_id: currentUser.id,
      organization_id: currentOrg?.id,
      start_date: form.start_date || undefined,
      end_date: form.end_date || undefined,
      status: form.status !== 'none' ? form.status : undefined,
    })
  }

  if (!currentUser) return null

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold text-gray-900">プロジェクト管理</h1>
          <p className="text-sm text-gray-500 mt-1">プロジェクトの作成・管理</p>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
        >
          <Plus className="w-4 h-4" />
          新規プロジェクト
        </button>
      </div>

      {isLoading ? (
        <div className="text-center py-16 text-gray-500">読み込み中...</div>
      ) : projects.length === 0 ? (
        <div className="text-center py-16 bg-white rounded-xl border border-dashed border-gray-300">
          <FolderKanban className="w-12 h-12 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-500">プロジェクトがありません</p>
          <button
            onClick={() => setShowForm(true)}
            className="mt-4 text-blue-600 text-sm font-medium hover:underline"
          >
            最初のプロジェクトを作成する
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {projects.map((project: Project) => (
            <Link
              key={project.id}
              href={`/admin/projects/${project.id}`}
              className="bg-white rounded-xl border border-gray-200 p-5 hover:border-blue-300 hover:shadow-md transition-all group flex items-center justify-between"
            >
              <div>
                <span className="text-xs font-mono font-bold text-blue-600 bg-blue-50 px-2 py-0.5 rounded">
                  {project.key}
                </span>
                <h3 className="mt-2 text-base font-semibold text-gray-900 group-hover:text-blue-600">
                  {project.name}
                </h3>
                {project.description && (
                  <p className="mt-1 text-sm text-gray-500 line-clamp-1">{project.description}</p>
                )}
                <div className="mt-2 flex items-center gap-2 text-xs text-gray-400">
                  <span>オーナー: {project.owner?.name}</span>
                  {project.status && project.status !== 'none' && (
                    <span className="px-1.5 py-0.5 rounded bg-gray-100 text-gray-600">
                      {PROJECT_STATUS_LABELS[project.status as ProjectStatus] ?? project.status}
                    </span>
                  )}
                </div>
              </div>
              <ChevronRight className="w-5 h-5 text-gray-400 group-hover:text-blue-500 flex-shrink-0 ml-4" />
            </Link>
          ))}
        </div>
      )}

      {/* 新規プロジェクト作成モーダル */}
      {showForm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl w-full max-w-md p-6 shadow-xl">
            <h3 className="text-lg font-bold text-gray-900 mb-4">新規プロジェクト作成</h3>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  プロジェクトキー <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={form.key}
                  onChange={(e) => setForm({ ...form, key: e.target.value.toUpperCase() })}
                  placeholder="例: PROJ"
                  maxLength={10}
                  required
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  プロジェクト名 <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="プロジェクト名を入力"
                  required
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">説明</label>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  placeholder="プロジェクトの説明（任意）"
                  rows={3}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">開始日</label>
                  <input
                    type="date"
                    value={form.start_date}
                    onChange={(e) => setForm({ ...form, start_date: e.target.value })}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">終了日</label>
                  <input
                    type="date"
                    value={form.end_date}
                    onChange={(e) => setForm({ ...form, end_date: e.target.value })}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">ライフサイクルステータス</label>
                <select
                  value={form.status}
                  onChange={(e) => setForm({ ...form, status: e.target.value as ProjectStatus })}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {(['none', 'planning', 'active', 'completed'] as const).map((s) => (
                    <option key={s} value={s}>
                      {PROJECT_STATUS_LABELS[s]}
                    </option>
                  ))}
                </select>
              </div>
              {createMutation.isError && (
                <p className="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-lg">作成に失敗しました</p>
              )}
              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg text-sm hover:bg-gray-50"
                >
                  キャンセル
                </button>
                <button
                  type="submit"
                  disabled={createMutation.isPending}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
                >
                  {createMutation.isPending ? '作成中...' : '作成する'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
