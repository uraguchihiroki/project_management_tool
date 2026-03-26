import axios, { isAxiosError } from 'axios'
import type {
  ApiResponse,
  ListResponse,
  Project,
  ProjectStatus,
  Issue,
  Comment,
  User,
  Status,
  IssueTemplate,
  Organization,
  SuperAdmin,
  Workflow,
  IssueEvent,
  WorkflowTransition,
} from '@/types'
import { clearAuthSession, getAuthToken } from '@/lib/authToken'

/**
 * API のベース URL（末尾スラッシュなし）。
 * - NEXT_PUBLIC_API_URL があれば最優先（本番・API を別ホストに置く場合）。
 * - 未設定かつブラウザでは同一オリジンの `/api/v1`（Next の rewrite → バックエンド）。
 *   → Windows 上の Playwright ブラウザから WSL の Next にアクセスしてもログイン API に届く。
 * - SSR / Node のフォールバックはループバック直叩き。
 */
export function resolveApiBaseURL(): string {
  const raw = process.env.NEXT_PUBLIC_API_URL
  if (typeof raw === 'string' && raw.trim() !== '') {
    return raw.replace(/\/+$/, '')
  }
  if (typeof window !== 'undefined' && window.location?.origin) {
    return `${window.location.origin.replace(/\/+$/, '')}/api/v1`
  }
  return 'http://127.0.0.1:8080/api/v1'
}

/** Next の rewrite 失敗や Echo の素の 500 など、プレーン「Internal Server Error」を判別 */
function isGenericInternalError(text: string): boolean {
  return /internal\s+server\s+error/i.test(text.trim())
}

/** 502/503/504 や上記のとき、ログイン画面向けに「APIに接続」系の文へ寄せる（E2E の期待文言とも整合） */
function messageForUnreachableOrBrokenApi(status: number, rawMessage: string): string | null {
  if ([502, 503, 504].includes(status)) {
    return 'ログインに失敗しました。APIに接続できません（プロキシ先のバックエンドが起動しているか確認してください）。'
  }
  if (status >= 500 && isGenericInternalError(rawMessage)) {
    return 'ログインに失敗しました。APIに接続できないか、サーバー内部エラーです（バックエンド起動・DATABASE_URL・Next の rewrite 先を確認してください）。'
  }
  return null
}

/** ログイン画面などで Axios 例外を人が読める文にする（Echo の message / ネットワークエラー等） */
export function formatApiError(e: unknown): string {
  if (isAxiosError(e)) {
    if (e.response) {
      const { status, statusText, data } = e.response
      if (typeof data === 'string' && data.trim()) {
        const t = data.trim().slice(0, 300)
        if (t.startsWith('<')) {
          return `HTTP ${status}（HTML が返りました。API の URL を確認してください）`
        }
        const mapped = messageForUnreachableOrBrokenApi(status, t)
        if (mapped) return mapped
        return t
      }
      if (data && typeof data === 'object' && 'message' in data) {
        const msg = String((data as { message: unknown }).message)
        const mapped = messageForUnreachableOrBrokenApi(status, msg)
        if (mapped) return mapped
        return msg
      }
      const fallback = `HTTP ${status}${statusText ? ` ${statusText}` : ''}`
      const mapped = messageForUnreachableOrBrokenApi(status, statusText || '')
      return mapped ?? fallback
    }
    if (e.code === 'ECONNABORTED') return '接続がタイムアウトしました'
    return e.message || 'サーバーに接続できません'
  }
  if (e instanceof Error) return e.message
  return String(e)
}

const api = axios.create({
  baseURL: resolveApiBaseURL(),
  headers: { 'Content-Type': 'application/json' },
  timeout: 30000,
})

api.interceptors.request.use((config) => {
  // モジュール読み込み時と異なる origin で動かすケースに備え毎回解決
  config.baseURL = resolveApiBaseURL()
  const token = getAuthToken()
  if (token) {
    config.headers = config.headers ?? {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (r) => r,
  (err) => {
    if (!axios.isAxiosError(err) || err.response?.status !== 401) {
      return Promise.reject(err)
    }
    const reqUrl = err.config?.url ?? ''
    if (
      reqUrl.includes('/admin/login') ||
      reqUrl.includes('/admin/switch-organization') ||
      reqUrl.includes('/super-admin/login')
    ) {
      return Promise.reject(err)
    }
    if (typeof window !== 'undefined') {
      clearAuthSession()
      const path = window.location.pathname
      if (path.startsWith('/super-admin')) {
        if (!path.startsWith('/super-admin/login')) {
          window.location.href = '/super-admin/login'
        }
      } else if (!path.startsWith('/login')) {
        window.location.href = '/login'
      }
    }
    return Promise.reject(err)
  }
)

// Users
export const getUsers = () =>
  api.get<ListResponse<User>>('/users').then((r) => r.data.data)

export const createUser = (data: { name: string; email: string }) =>
  api.post<ApiResponse<User>>('/users', data).then((r) => r.data.data)

/** メールのみログイン（JWT 発行）。未認証で呼ぶ。 */
export const adminLogin = (email: string) =>
  api
    .post<ApiResponse<{ user: User; token: string }>>('/admin/login', { email })
    .then((r) => r.data.data)

/** JWT の組織スコープを切り替え（同一メールの別組織ユーザー行に紐づくトークンを再発行） */
export const switchOrganization = (organizationId: string) =>
  api
    .post<ApiResponse<{ user: User; token: string }>>('/admin/switch-organization', {
      organization_id: organizationId,
    })
    .then((r) => r.data.data)

export const setUserAdmin = (userId: string, isAdmin: boolean) =>
  api.put<ApiResponse<unknown>>(`/users/${userId}/admin`, { is_admin: isAdmin })

// Admin users (org-scoped)
export const getAdminUsers = (orgId: string) =>
  api.get<ListResponse<User>>('/admin/users', { params: { org_id: orgId } }).then((r) => r.data.data)

export const createAdminUser = (orgId: string, name: string, email: string) =>
  api.post<ApiResponse<User>>('/admin/users', { org_id: orgId, name, email }).then((r) => r.data.data)

export const updateAdminUser = (id: string, name: string) =>
  api.put<ApiResponse<unknown>>(`/admin/users/${id}`, { name }).then((r) => r.data)

export const deleteAdminUser = (id: string, orgId: string) =>
  api.delete(`/admin/users/${id}`, { params: { org_id: orgId } })

// Organizations
export const getOrganizations = () =>
  api.get<ListResponse<Organization>>('/organizations').then((r) => r.data.data)

export const createOrganization = (name: string) =>
  api.post<ApiResponse<Organization>>('/organizations', { name }).then((r) => r.data.data)

export const getUserOrganizations = (userId: string) =>
  api.get<{ data: Organization[] }>(`/users/${userId}/organizations`).then((r) => r.data.data)

export const addUserToOrg = (orgId: string, userId: string, isOrgAdmin = false) =>
  api.post(`/organizations/${orgId}/users`, { user_id: userId, is_org_admin: isOrgAdmin })

// Super Admin
export const superAdminLogin = (email: string) =>
  api.post<ApiResponse<SuperAdmin>>('/super-admin/login', { email }).then((r) => r.data.data)

export const superAdminGetOrganizations = () =>
  api.get<ListResponse<Organization>>('/super-admin/organizations').then((r) => r.data.data)

export const superAdminCreateOrganization = (data: { name: string; admin_email?: string; admin_name?: string }) =>
  api.post<ApiResponse<Organization>>('/super-admin/organizations', data).then((r) => r.data.data)

// Projects
export const getProjects = (orgId?: string) =>
  api.get<ListResponse<Project>>('/projects', { params: orgId ? { org_id: orgId } : {} }).then((r) => r.data.data)

export const getProject = (id: string) =>
  api.get<ApiResponse<Project>>(`/projects/${id}`).then((r) => r.data.data)

export const createProject = (data: {
  key: string
  name: string
  description?: string
  owner_id: string
  organization_id?: string
  start_date?: string
  end_date?: string
}) => api.post<ApiResponse<Project>>('/projects', data).then((r) => r.data.data)

/** 未設定時のみデフォルト Issue 用ワークフローとカンバン列を紐付ける（冪等） */
export const ensureDefaultIssueWorkflow = (projectId: string) =>
  api.post<ApiResponse<Project>>(`/projects/${projectId}/default-issue-workflow`).then((r) => r.data.data)

export const updateProject = (
  id: string,
  data: {
    name?: string
    description?: string
    start_date?: string
    end_date?: string
    project_status_id?: string
  }
) => api.put<ApiResponse<Project>>(`/projects/${id}`, data).then((r) => r.data.data)

export const deleteProject = (id: string) =>
  api.delete(`/projects/${id}`)

// Statuses
export const getProjectStatuses = (projectId: string) =>
  api.get<ListResponse<ProjectStatus>>(`/projects/${projectId}/project-statuses`).then((r) => r.data.data)

export const updateProjectStatus = (
  projectId: string,
  statusId: string,
  data: { name: string; color: string; order: number }
) =>
  api
    .put<ApiResponse<ProjectStatus>>(`/projects/${projectId}/project-statuses/${statusId}`, data)
    .then((r) => r.data.data)

// Issues
export const getIssues = (projectId: string) =>
  api.get<ListResponse<Issue>>(`/projects/${projectId}/issues`).then((r) => r.data.data)

export const getIssue = (projectId: string, number: number) =>
  api.get<ApiResponse<Issue>>(`/projects/${projectId}/issues/${number}`).then((r) => r.data.data)

export const createIssue = (
  projectId: string,
  data: {
    title: string
    description?: string
    status_id: string
    priority?: string
    assignee_id?: string
    reporter_id: string
    due_date?: string
    template_id?: number
  }
) => api.post<ApiResponse<Issue>>(`/projects/${projectId}/issues`, data).then((r) => r.data.data)

// Templates
export const getTemplates = () =>
  api.get<ListResponse<IssueTemplate>>('/templates').then((r) => r.data.data)

export const getProjectTemplates = (projectId: string) =>
  api.get<ListResponse<IssueTemplate>>(`/projects/${projectId}/templates`).then((r) => r.data.data)

export const createTemplate = (data: {
  project_id: string
  name: string
  description?: string
  body?: string
  default_priority?: string
}) => api.post<ApiResponse<IssueTemplate>>('/templates', data).then((r) => r.data.data)

export const updateTemplate = (id: number, data: {
  name: string
  description?: string
  body?: string
  default_priority?: string
}) => api.put<ApiResponse<IssueTemplate>>(`/templates/${id}`, data).then((r) => r.data.data)

export const deleteTemplate = (id: number) =>
  api.delete(`/templates/${id}`)

export const updateIssue = (
  projectId: string,
  number: number,
  data: Partial<{
    title: string
    description: string
    status_id: string
    priority: string
    assignee_id: string
  }>
) => api.put<ApiResponse<Issue>>(`/projects/${projectId}/issues/${number}`, data).then((r) => r.data.data)

/** Issue のインプリント（時系列） */
export const getIssueEvents = (issueId: string) =>
  api.get<{ data: IssueEvent[] }>(`/issues/${issueId}/events`).then((r) => r.data.data)

export const deleteIssue = (projectId: string, number: number) =>
  api.delete(`/projects/${projectId}/issues/${number}`)

// Workflows
/** 管理画面で選択中の組織に合わせる場合は orgId を渡す（スーパーアドミン時のサーバー側絞り込み） */
export const getWorkflows = (orgId?: string) =>
  api
    .get<ListResponse<Workflow>>('/workflows', {
      params: orgId ? { org_id: orgId } : undefined,
    })
    .then((r) => r.data.data)

export const createWorkflow = (data: {
  organization_id: string
  name: string
  description?: string
}) => api.post<ApiResponse<Workflow>>('/workflows', data).then((r) => r.data.data)

export const updateWorkflowMeta = (id: string | number, data: { name: string; description: string }) =>
  api.put<ApiResponse<Workflow>>(`/workflows/${id}`, data).then((r) => r.data.data)

/** 管理画面: ワークフロー名・説明・ステータス・許可遷移の一括保存（204 No Content） */
export type WorkflowEditorStatusPayload = {
  id?: string
  client_id?: string
  name: string
  color: string
  is_entry: boolean
  is_terminal: boolean
}

export type WorkflowEditorTransitionPayload = {
  from_ref: string
  to_ref: string
}

export const saveWorkflowEditor = (
  workflowId: string,
  data: {
    name: string
    description: string
    statuses: WorkflowEditorStatusPayload[]
    transitions: WorkflowEditorTransitionPayload[]
  }
) => api.put(`/workflows/${workflowId}/editor`, data).then(() => undefined)

export const deleteWorkflowApi = (id: string | number) => api.delete(`/workflows/${id}`)

export const reorderWorkflowsApi = (ids: number[]) =>
  api.put('/workflows/reorder', { ids }).then(() => undefined)

export const getWorkflow = (id: string) =>
  api.get<ApiResponse<Workflow>>(`/workflows/${id}`).then((r) => r.data.data)

export const getWorkflowStatuses = (workflowId: string) =>
  api.get<ListResponse<Status>>(`/workflows/${workflowId}/statuses`).then((r) => r.data.data)

export const getWorkflowTransitions = (workflowId: string) =>
  api.get<ListResponse<WorkflowTransition>>(`/workflows/${workflowId}/transitions`).then((r) => r.data.data)

export const createWorkflowTransition = (
  workflowId: string,
  data: { from_status_id: string; to_status_id: string }
) =>
  api
    .post<ApiResponse<WorkflowTransition>>(`/workflows/${workflowId}/transitions`, data)
    .then((r) => r.data.data)

export const deleteWorkflowTransition = (workflowId: string, transitionId: number) =>
  api.delete(`/workflows/${workflowId}/transitions/${transitionId}`).then(() => undefined)

export const updateWorkflowTransition = (
  workflowId: string,
  transitionId: number,
  data: { from_status_id: string; to_status_id: string }
) =>
  api
    .put<ApiResponse<WorkflowTransition>>(`/workflows/${workflowId}/transitions/${transitionId}`, data)
    .then((r) => r.data.data)

export const createWorkflowStatus = (
  workflowId: string,
  data: { name: string; color?: string; display_order?: number }
) =>
  api.post<ApiResponse<Status>>(`/workflows/${workflowId}/statuses`, data).then((r) => r.data.data)

export const updateStatus = (
  id: string,
  data: {
    name?: string
    color?: string
    display_order: number
    is_entry?: boolean
    is_terminal?: boolean
  }
) => api.put<ApiResponse<Status>>(`/statuses/${id}`, data).then((r) => r.data.data)

export const reorderWorkflowStatuses = (workflowId: string, statusIds: string[]) =>
  api.put(`/workflows/${workflowId}/statuses/reorder`, { status_ids: statusIds }).then(() => undefined)

export const reorderWorkflowTransitions = (workflowId: string, transitionIds: number[]) =>
  api.put(`/workflows/${workflowId}/transitions/reorder`, { transition_ids: transitionIds }).then(() => undefined)

export const deleteStatus = (id: string) =>
  api.delete(`/statuses/${id}`).then(() => undefined)

// Comments
export const getComments = (issueId: string) =>
  api.get<ListResponse<Comment>>(`/issues/${issueId}/comments`).then((r) => r.data.data)

export const createComment = (issueId: string, data: { author_id: string; body: string }) =>
  api.post<ApiResponse<Comment>>(`/issues/${issueId}/comments`, data).then((r) => r.data.data)

export const deleteComment = (issueId: string, commentId: string) =>
  api.delete(`/issues/${issueId}/comments/${commentId}`)
