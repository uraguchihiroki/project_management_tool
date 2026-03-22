import axios from 'axios'
import type {
  ApiResponse,
  ListResponse,
  Project,
  Issue,
  Comment,
  User,
  Status,
  IssueTemplate,
  Organization,
  SuperAdmin,
  Workflow,
  Group,
  IssueEvent,
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

const api = axios.create({
  baseURL: resolveApiBaseURL(),
  headers: { 'Content-Type': 'application/json' },
  timeout: 30000,
})

api.interceptors.request.use((config) => {
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

export const updateProject = (
  id: string,
  data: { name?: string; description?: string; start_date?: string; end_date?: string }
) => api.put<ApiResponse<Project>>(`/projects/${id}`, data).then((r) => r.data.data)

export const deleteProject = (id: string) =>
  api.delete(`/projects/${id}`)

// Statuses
export const getStatuses = (projectId: string) =>
  api.get<ListResponse<Status>>(`/projects/${projectId}/statuses`).then((r) => r.data.data)

// Issues
export const getIssues = (projectId: string, opts?: { group_id?: string }) =>
  api
    .get<ListResponse<Issue>>(`/projects/${projectId}/issues`, {
      params: opts?.group_id ? { group_id: opts.group_id } : {},
    })
    .then((r) => r.data.data)

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
    group_ids?: string[]
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
    group_ids: string[]
  }>
) => api.put<ApiResponse<Issue>>(`/projects/${projectId}/issues/${number}`, data).then((r) => r.data.data)

/** Issue のインプリント（時系列） */
export const getIssueEvents = (issueId: string) =>
  api.get<{ data: IssueEvent[] }>(`/issues/${issueId}/events`).then((r) => r.data.data)

/** 組織のグループ一覧 */
export const getOrganizationGroups = (orgId: string, params?: { kind?: string }) =>
  api.get<ListResponse<Group>>(`/organizations/${orgId}/groups`, { params: params ?? {} }).then((r) => r.data.data)

/** Issue に付いたグループ一覧（GET は Issue 詳細に含まれることもあるが、明示取得用） */
export const getIssueGroups = (projectId: string, number: number) =>
  api.get<{ data: Group[] }>(`/projects/${projectId}/issues/${number}/groups`).then((r) => r.data.data)

export const putIssueGroups = (projectId: string, number: number, groupIds: string[]) =>
  api.put(`/projects/${projectId}/issues/${number}/groups`, { group_ids: groupIds })

export const getUserGroups = (userId: string) =>
  api.get<ListResponse<Group>>(`/users/${userId}/groups`).then((r) => r.data.data)

export const deleteIssue = (projectId: string, number: number) =>
  api.delete(`/projects/${projectId}/issues/${number}`)

// Workflows
export const getWorkflows = () =>
  api.get<ListResponse<Workflow>>('/workflows').then((r) => r.data.data)

export const createWorkflow = (data: {
  organization_id: string
  name: string
  description?: string
}) => api.post<ApiResponse<Workflow>>('/workflows', data).then((r) => r.data.data)

export const updateWorkflowMeta = (id: string | number, data: { name: string; description: string }) =>
  api.put<ApiResponse<Workflow>>(`/workflows/${id}`, data).then((r) => r.data.data)

export const deleteWorkflowApi = (id: string | number) => api.delete(`/workflows/${id}`)

export const reorderWorkflowsApi = (ids: number[]) =>
  api.put('/workflows/reorder', { ids }).then(() => undefined)

export const getWorkflow = (id: string) =>
  api.get<ApiResponse<Workflow>>(`/workflows/${id}`).then((r) => r.data.data)

// Comments
export const getComments = (issueId: string) =>
  api.get<ListResponse<Comment>>(`/issues/${issueId}/comments`).then((r) => r.data.data)

export const createComment = (issueId: string, data: { author_id: string; body: string }) =>
  api.post<ApiResponse<Comment>>(`/issues/${issueId}/comments`, data).then((r) => r.data.data)

export const deleteComment = (issueId: string, commentId: string) =>
  api.delete(`/issues/${issueId}/comments/${commentId}`)
