'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Shield, ShieldOff, X, Check, Plus, Pencil, Trash2 } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'
import { getAdminUsers, createAdminUser, updateAdminUser, deleteAdminUser } from '@/lib/api'
import type { Role, User } from '@/types'

const API = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

async function fetchRoles(orgId?: string): Promise<Role[]> {
  const url = orgId ? `${API}/roles?org_id=${orgId}` : `${API}/roles`
  const res = await fetch(url)
  const json = await res.json()
  return json.data ?? []
}

export default function AdminUsersPage() {
  const { currentOrg } = useAuth()
  const queryClient = useQueryClient()
  const { data: users = [], isLoading } = useQuery({
    queryKey: ['admin-users', currentOrg?.id],
    queryFn: () => getAdminUsers(currentOrg!.id),
    enabled: !!currentOrg?.id,
  })
  const { data: allRoles = [] } = useQuery({
    queryKey: ['roles', currentOrg?.id],
    queryFn: () => fetchRoles(currentOrg?.id),
    enabled: !!currentOrg?.id,
  })

  const [showCreateForm, setShowCreateForm] = useState(false)
  const [createForm, setCreateForm] = useState({ name: '', email: '' })
  const [editingNameUserId, setEditingNameUserId] = useState<string | null>(null)
  const [editingRoleUserId, setEditingRoleUserId] = useState<string | null>(null)
  const [editingName, setEditingName] = useState('')
  const [selectedRoleIds, setSelectedRoleIds] = useState<number[]>([])

  const createMutation = useMutation({
    mutationFn: () => createAdminUser(currentOrg!.id, createForm.name, createForm.email),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setShowCreateForm(false)
      setCreateForm({ name: '', email: '' })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ userId, name }: { userId: string; name: string }) => updateAdminUser(userId, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setEditingNameUserId(null)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (userId: string) => deleteAdminUser(userId, currentOrg!.id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-users'] }),
  })

  const setAdminMutation = useMutation({
    mutationFn: async ({ userId, isAdmin }: { userId: string; isAdmin: boolean }) => {
      const res = await fetch(`${API}/users/${userId}/admin`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ is_admin: isAdmin }),
      })
      if (!res.ok) throw new Error('更新に失敗しました')
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-users'] }),
  })

  const assignRolesMutation = useMutation({
    mutationFn: async ({ userId, roleIds }: { userId: string; roleIds: number[] }) => {
      const res = await fetch(`${API}/users/${userId}/roles`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role_ids: roleIds }),
      })
      if (!res.ok) throw new Error('役職の更新に失敗しました')
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setEditingRoleUserId(null)
    },
  })

  const startRoleEdit = (user: User) => {
    setEditingRoleUserId(user.id)
    setSelectedRoleIds((user.roles ?? []).map((r) => r.id))
  }

  const startNameEdit = (user: User) => {
    setEditingNameUserId(user.id)
    setEditingName(user.name)
  }

  const toggleRole = (roleId: number) => {
    setSelectedRoleIds((prev) =>
      prev.includes(roleId) ? prev.filter((id) => id !== roleId) : [...prev, roleId]
    )
  }

  if (!currentOrg) {
    return (
      <div className="p-8 text-center text-gray-500">
        組織を選択してください
      </div>
    )
  }

  return (
    <div className="max-w-4xl">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900">ユーザー管理</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            {currentOrg.name} のユーザーを管理します（作成・更新・削除・役職）
          </p>
        </div>
        <button
          onClick={() => setShowCreateForm(!showCreateForm)}
          className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          <Plus className="w-4 h-4" />
          ユーザー登録
        </button>
      </div>

      <div className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden">
        {showCreateForm && (
          <div className="px-6 py-4 border-b border-gray-100 bg-gray-50">
            <form
              onSubmit={(e) => {
                e.preventDefault()
                if (createForm.name && createForm.email) createMutation.mutate()
              }}
              className="flex gap-3 flex-wrap items-end"
            >
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">名前</label>
                <input
                  type="text"
                  value={createForm.name}
                  onChange={(e) => setCreateForm((f) => ({ ...f, name: e.target.value }))}
                  placeholder="山田 太郎"
                  className="w-40 px-3 py-2 border border-gray-300 rounded-lg text-sm"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">メール *</label>
                <input
                  type="email"
                  value={createForm.email}
                  onChange={(e) => setCreateForm((f) => ({ ...f, email: e.target.value }))}
                  placeholder="taro@example.com"
                  required
                  className="w-48 px-3 py-2 border border-gray-300 rounded-lg text-sm"
                />
              </div>
              <div className="flex gap-2">
                <button
                  type="submit"
                  disabled={createMutation.isPending || !createForm.email}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700 disabled:opacity-50"
                >
                  登録
                </button>
                <button
                  type="button"
                  onClick={() => setShowCreateForm(false)}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-sm hover:bg-gray-50"
                >
                  キャンセル
                </button>
              </div>
            </form>
          </div>
        )}

        {isLoading ? (
          <div className="p-8 text-center text-gray-400 text-sm">読み込み中...</div>
        ) : users.length === 0 ? (
          <div className="p-8 text-center text-gray-400 text-sm">ユーザーがいません</div>
        ) : (
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide">ユーザー</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide">役職</th>
                <th className="text-center px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide w-24">管理者</th>
                <th className="px-4 py-3 w-32"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {users.map((user) => (
                <tr key={user.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3">
                    {editingNameUserId === user.id ? (
                      <div className="flex items-center gap-2">
                        <input
                          type="text"
                          value={editingName}
                          onChange={(e) => setEditingName(e.target.value)}
                          className="w-32 px-2 py-1 border border-gray-300 rounded text-sm"
                          autoFocus
                        />
                        <button
                          onClick={() => updateMutation.mutate({ userId: user.id, name: editingName })}
                          disabled={updateMutation.isPending}
                          className="p-1 text-blue-600 hover:bg-blue-50 rounded"
                        >
                          <Check className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => setEditingNameUserId(null)}
                          className="p-1 text-gray-400 hover:bg-gray-100 rounded"
                        >
                          <X className="w-4 h-4" />
                        </button>
                      </div>
                    ) : (
                      <div className="flex items-center gap-2 group">
                        <div>
                          <p className="font-medium text-gray-900 text-sm">{user.name}</p>
                          <p className="text-xs text-gray-400">{user.email}</p>
                        </div>
                        <button
                          onClick={() => startNameEdit(user)}
                          className="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-blue-600 rounded"
                          title="名前を編集"
                        >
                          <Pencil className="w-3 h-3" />
                        </button>
                      </div>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    {editingRoleUserId === user.id ? (
                      <div className="space-y-2">
                        <div className="flex flex-wrap gap-1.5">
                          {allRoles.length === 0 ? (
                            <span className="text-xs text-gray-400">役職がありません</span>
                          ) : (
                            allRoles.map((role) => (
                              <button
                                key={role.id}
                                onClick={() => toggleRole(role.id)}
                                className={`flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium border transition-colors ${
                                  selectedRoleIds.includes(role.id)
                                    ? 'bg-blue-100 text-blue-700 border-blue-300'
                                    : 'bg-white text-gray-500 border-gray-300 hover:border-blue-300'
                                }`}
                              >
                                <span>Lv.{role.level}</span>
                                <span>{role.name}</span>
                              </button>
                            ))
                          )}
                        </div>
                        <div className="flex gap-1.5">
                          <button
                            onClick={() => assignRolesMutation.mutate({ userId: user.id, roleIds: selectedRoleIds })}
                            disabled={assignRolesMutation.isPending}
                            className="flex items-center gap-1 px-2.5 py-1 bg-blue-600 text-white rounded-md text-xs font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
                          >
                            <Check className="w-3 h-3" />
                            保存
                          </button>
                          <button
                            onClick={() => setEditingRoleUserId(null)}
                            className="flex items-center gap-1 px-2.5 py-1 border border-gray-300 text-gray-600 rounded-md text-xs hover:bg-gray-50 transition-colors"
                          >
                            <X className="w-3 h-3" />
                            キャンセル
                          </button>
                        </div>
                      </div>
                    ) : editingRoleUserId !== user.id ? (
                      <button
                        onClick={() => startRoleEdit(user)}
                        className="flex flex-wrap gap-1 group"
                        title="クリックして役職を編集"
                      >
                        {(user.roles ?? []).length === 0 ? (
                          <span className="text-xs text-gray-400 group-hover:text-blue-500 transition-colors">
                            役職なし（クリックして設定）
                          </span>
                        ) : (
                          (user.roles ?? []).map((role) => (
                            <span
                              key={role.id}
                              className="inline-flex items-center gap-1 px-2 py-0.5 bg-blue-50 text-blue-700 rounded-md text-xs font-medium"
                            >
                              <span className="text-blue-400">Lv.{role.level}</span>
                              {role.name}
                            </span>
                          ))
                        )}
                      </button>
                    ) : null}
                  </td>
                  <td className="px-4 py-3 text-center">
                    <button
                      onClick={() => {
                        if (confirm(`${user.name}の管理者フラグを${user.is_admin ? '解除' : '付与'}しますか？`)) {
                          setAdminMutation.mutate({ userId: user.id, isAdmin: !user.is_admin })
                        }
                      }}
                      className={`inline-flex items-center justify-center w-8 h-8 rounded-full transition-colors ${
                        user.is_admin
                          ? 'bg-blue-100 text-blue-600 hover:bg-red-100 hover:text-red-600'
                          : 'bg-gray-100 text-gray-400 hover:bg-blue-100 hover:text-blue-600'
                      }`}
                      title={user.is_admin ? '管理者を解除' : '管理者に昇格'}
                    >
                      {user.is_admin ? <Shield className="w-4 h-4" /> : <ShieldOff className="w-4 h-4" />}
                    </button>
                  </td>
                  <td className="px-4 py-3">
                    {editingNameUserId !== user.id && editingRoleUserId !== user.id && (
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => startRoleEdit(user)}
                          className="text-xs text-blue-600 hover:underline"
                        >
                          役職編集
                        </button>
                        <button
                          onClick={() => {
                            if (confirm(`${user.name} をこの組織から除外しますか？`)) {
                              deleteMutation.mutate(user.id)
                            }
                          }}
                          className="p-1 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded"
                          title="組織から除外"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
