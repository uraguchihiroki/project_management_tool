'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Shield, ShieldOff, X, Check, Plus, Pencil, Trash2 } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'
import { getAdminUsers, createAdminUser, updateAdminUser, deleteAdminUser, resolveApiBaseURL } from '@/lib/api'
import { useAuthFetchEnabled } from '@/hooks/useAuthFetchEnabled'
import type { User, Group } from '@/types'

async function fetchGroups(orgId: string): Promise<Group[]> {
  const res = await fetch(`${resolveApiBaseURL()}/organizations/${orgId}/groups`)
  const json = await res.json()
  return json.data ?? []
}

async function fetchUserGroups(orgId: string, userId: string): Promise<Group[]> {
  const res = await fetch(`${resolveApiBaseURL()}/users/${userId}/groups?org_id=${orgId}`)
  const json = await res.json()
  return json.data ?? []
}

export default function AdminUsersPage() {
  const { currentOrg } = useAuth()
  const authFetch = useAuthFetchEnabled()
  const queryClient = useQueryClient()
  const { data: users = [], isLoading } = useQuery({
    queryKey: ['admin-users', currentOrg?.id],
    queryFn: () => getAdminUsers(currentOrg!.id),
    enabled: authFetch && !!currentOrg?.id,
  })
  const { data: allGroups = [] } = useQuery({
    queryKey: ['groups', currentOrg?.id],
    queryFn: () => fetchGroups(currentOrg!.id),
    enabled: authFetch && !!currentOrg?.id,
  })

  const [showCreateForm, setShowCreateForm] = useState(false)
  const [createForm, setCreateForm] = useState({ name: '', email: '' })
  const [editingNameUserId, setEditingNameUserId] = useState<string | null>(null)
  const [editingDeptUserId, setEditingDeptUserId] = useState<string | null>(null)
  const [editingName, setEditingName] = useState('')
  const [selectedGroupIds, setSelectedGroupIds] = useState<string[]>([])
  const [userGroupsCache, setUserGroupsCache] = useState<Record<string, Group[]>>({})

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
      const res = await fetch(`${resolveApiBaseURL()}/users/${userId}/admin`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ is_admin: isAdmin }),
      })
      if (!res.ok) throw new Error('更新に失敗しました')
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['admin-users'] }),
  })

  const assignDepartmentsMutation = useMutation({
    mutationFn: async ({ userId, groupIds }: { userId: string; groupIds: string[] }) => {
      const res = await fetch(`${resolveApiBaseURL()}/users/${userId}/groups?org_id=${currentOrg!.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ group_ids: groupIds }),
      })
      if (!res.ok) throw new Error('グループの更新に失敗しました')
    },
    onSuccess: (_, { userId, groupIds }) => {
      const groups = allGroups.filter((g) => groupIds.includes(g.id))
      setUserGroupsCache((c) => ({ ...c, [userId]: groups }))
      queryClient.invalidateQueries({ queryKey: ['admin-users'] })
      setEditingDeptUserId(null)
    },
  })

  const startDeptEdit = async (user: User) => {
    setEditingDeptUserId(user.id)
    const groups = await fetchUserGroups(currentOrg!.id, user.id)
    setSelectedGroupIds(groups.map((g) => g.id))
    setUserGroupsCache((c) => ({ ...c, [user.id]: groups }))
  }

  const startNameEdit = (user: User) => {
    setEditingNameUserId(user.id)
    setEditingName(user.name)
  }

  const toggleGroup = (groupId: string) => {
    setSelectedGroupIds((prev) =>
      prev.includes(groupId) ? prev.filter((id) => id !== groupId) : [...prev, groupId]
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
            {currentOrg.name} のユーザーを管理します（作成・更新・削除・グループ）
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
                <th className="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase tracking-wide">グループ</th>
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
                    {editingDeptUserId === user.id ? (
                      <div className="space-y-2">
                        <div className="flex flex-wrap gap-1.5">
                          {allGroups.length === 0 ? (
                            <span className="text-xs text-gray-400">グループがありません</span>
                          ) : (
                            allGroups.map((group) => (
                              <button
                                key={group.id}
                                onClick={() => toggleGroup(group.id)}
                                className={`flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium border transition-colors ${
                                  selectedGroupIds.includes(group.id)
                                    ? 'bg-green-100 text-green-700 border-green-300'
                                    : 'bg-white text-gray-500 border-gray-300 hover:border-green-300'
                                }`}
                              >
                                {group.name}
                              </button>
                            ))
                          )}
                        </div>
                        <div className="flex gap-1.5">
                          <button
                            onClick={() => assignDepartmentsMutation.mutate({ userId: user.id, groupIds: selectedGroupIds })}
                            disabled={assignDepartmentsMutation.isPending}
                            className="flex items-center gap-1 px-2.5 py-1 bg-blue-600 text-white rounded-md text-xs font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
                          >
                            <Check className="w-3 h-3" />
                            保存
                          </button>
                          <button
                            onClick={() => setEditingDeptUserId(null)}
                            className="flex items-center gap-1 px-2.5 py-1 border border-gray-300 text-gray-600 rounded-md text-xs hover:bg-gray-50 transition-colors"
                          >
                            <X className="w-3 h-3" />
                            キャンセル
                          </button>
                        </div>
                      </div>
                    ) : (
                      <button
                        onClick={() => startDeptEdit(user)}
                        className="flex flex-wrap gap-1 group"
                        title="クリックしてグループを編集"
                      >
                        {(userGroupsCache[user.id] ?? []).length === 0 && editingDeptUserId !== user.id ? (
                          <span className="text-xs text-gray-400 group-hover:text-blue-500 transition-colors">
                            グループなし（クリックして設定）
                          </span>
                        ) : (
                          (userGroupsCache[user.id] ?? []).map((group) => (
                            <span
                              key={group.id}
                              className="inline-flex items-center gap-1 px-2 py-0.5 bg-green-50 text-green-700 rounded-md text-xs font-medium"
                            >
                              {group.name}
                            </span>
                          ))
                        )}
                      </button>
                    )}
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
                    {editingNameUserId !== user.id && editingDeptUserId !== user.id && (
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => startDeptEdit(user)}
                          className="text-xs text-green-600 hover:underline"
                        >
                          グループ編集
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
