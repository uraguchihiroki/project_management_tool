'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getProjects, createProject, getUsers, createUser } from '@/lib/api'
import { useState } from 'react'
import Link from 'next/link'
import { Plus, FolderKanban, ChevronRight, User } from 'lucide-react'
import type { Project } from '@/types'

export default function ProjectsPage() {
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [showUserForm, setShowUserForm] = useState(false)
  const [form, setForm] = useState({ key: '', name: '', description: '' })
  const [userForm, setUserForm] = useState({ name: '', email: '' })

  const { data: projects = [], isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: getProjects,
  })

  const { data: users = [] } = useQuery({
    queryKey: ['users'],
    queryFn: getUsers,
  })

  const createUserMutation = useMutation({
    mutationFn: (data: { name: string; email: string }) => createUser(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      setShowUserForm(false)
      setUserForm({ name: '', email: '' })
    },
  })

  const createMutation = useMutation({
    mutationFn: (data: Parameters<typeof createProject>[0]) => createProject(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      setShowForm(false)
      setForm({ key: '', name: '', description: '' })
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!users[0]) {
      setShowForm(false)
      setShowUserForm(true)
      return
    }
    createMutation.mutate({
      key: form.key.toUpperCase(),
      name: form.name,
      description: form.description || undefined,
      owner_id: users[0].id,
    })
  }

  const handleUserSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    createUserMutation.mutate(userForm)
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="max-w-6xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-2">
            <FolderKanban className="w-6 h-6 text-blue-600" />
            <h1 className="text-xl font-bold text-gray-900">ProjectHub</h1>
          </div>
          <div className="flex items-center gap-3">
            {users[0] ? (
              <div className="flex items-center gap-2 text-sm text-gray-600 bg-gray-100 px-3 py-1.5 rounded-lg">
                <User className="w-4 h-4" />
                <span>{users[0].name}</span>
              </div>
            ) : (
              <button
                onClick={() => setShowUserForm(true)}
                className="flex items-center gap-2 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg text-sm font-medium hover:bg-gray-50 transition-colors"
              >
                <User className="w-4 h-4" />
                ユーザー登録
              </button>
            )}
            <button
              onClick={() => setShowForm(true)}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
            >
              <Plus className="w-4 h-4" />
              新規プロジェクト
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-6 py-8">
        <h2 className="text-2xl font-bold text-gray-900 mb-6">プロジェクト一覧</h2>

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

      {/* User Registration Modal */}
      {showUserForm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl w-full max-w-md p-6 shadow-xl">
            <h3 className="text-lg font-bold text-gray-900 mb-1">ユーザー登録</h3>
            <p className="text-sm text-gray-500 mb-4">プロジェクト作成前に、まずユーザーを登録してください。</p>
            <form onSubmit={handleUserSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  名前 <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={userForm.name}
                  onChange={(e) => setUserForm({ ...userForm, name: e.target.value })}
                  placeholder="例: 山田 太郎"
                  required
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  メールアドレス <span className="text-red-500">*</span>
                </label>
                <input
                  type="email"
                  value={userForm.email}
                  onChange={(e) => setUserForm({ ...userForm, email: e.target.value })}
                  placeholder="例: taro@example.com"
                  required
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowUserForm(false)}
                  className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg text-sm hover:bg-gray-50"
                >
                  キャンセル
                </button>
                <button
                  type="submit"
                  disabled={createUserMutation.isPending}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
                >
                  {createUserMutation.isPending ? '登録中...' : '登録する'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Project Creation Modal */}
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
