'use client'

import { use } from 'react'
import Link from 'next/link'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getProject, updateProject } from '@/lib/api'
import { ChevronLeft } from 'lucide-react'
import type { Project } from '@/types'
import { useRequireAdmin } from '@/context/AuthContext'

export default function AdminProjectEditPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const currentUser = useRequireAdmin()
  const queryClient = useQueryClient()

  const { data: project, isLoading } = useQuery({
    queryKey: ['project', id],
    queryFn: () => getProject(id),
  })

  const updateMutation = useMutation({
    mutationFn: (data: Parameters<typeof updateProject>[1]) => updateProject(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', id] })
      queryClient.invalidateQueries({ queryKey: ['projects'] })
    },
  })

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!currentUser || !project) return
    const form = e.currentTarget
    const formData = new FormData(form)
    updateMutation.mutate({
      name: (formData.get('name') as string) || undefined,
      description: (formData.get('description') as string) || undefined,
      start_date: (formData.get('start_date') as string) || undefined,
      end_date: (formData.get('end_date') as string) || undefined,
    })
  }

  if (!currentUser) return null
  if (isLoading) return <div className="text-center py-16 text-gray-500">読み込み中...</div>
  if (!project) return <div className="text-center py-16 text-gray-500">プロジェクトが見つかりません</div>

  return (
    <div>
      <Link
        href="/admin/projects"
        className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-blue-600 mb-6"
      >
        <ChevronLeft className="w-4 h-4" />
        プロジェクト管理に戻る
      </Link>

      <h1 className="text-xl font-bold text-gray-900 mb-6">プロジェクト編集</h1>

      <form onSubmit={handleSubmit} className="max-w-xl space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">プロジェクトキー</label>
          <input
            type="text"
            value={project.key}
            disabled
            className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm bg-gray-50 text-gray-500"
          />
          <p className="mt-0.5 text-xs text-gray-400">キーは変更できません</p>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            プロジェクト名 <span className="text-red-500">*</span>
          </label>
          <input
            name="name"
            type="text"
            defaultValue={project.name}
            required
            className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">説明</label>
          <textarea
            name="description"
            defaultValue={project.description ?? ''}
            rows={3}
            className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">開始日</label>
            <input
              name="start_date"
              type="date"
              defaultValue={project.start_date ?? ''}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">終了日</label>
            <input
              name="end_date"
              type="date"
              defaultValue={project.end_date ?? ''}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>

        {updateMutation.isError && (
          <p className="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-lg">
            {updateMutation.error instanceof Error ? updateMutation.error.message : '更新に失敗しました'}
          </p>
        )}

        <div className="flex gap-3 pt-2">
          <Link
            href="/admin/projects"
            className="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg text-sm hover:bg-gray-50"
          >
            キャンセル
          </Link>
          <button
            type="submit"
            disabled={updateMutation.isPending}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
          >
            {updateMutation.isPending ? '保存中...' : '保存'}
          </button>
        </div>
      </form>
    </div>
  )
}
