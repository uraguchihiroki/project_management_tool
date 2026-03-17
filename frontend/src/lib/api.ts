import axios from 'axios'
import type { ApiResponse, ListResponse, Project, Issue, Comment, User, Status, IssueTemplate, IssueApproval, Organization, SuperAdmin } from '@/types'

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

// Users
export const getUsers = () =>
  api.get<ListResponse<User>>('/users').then((r) => r.data.data)

export const createUser = (data: { name: string; email: string }) =>
  api.post<ApiResponse<User>>('/users', data).then((r) => r.data.data)

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

export const superAdminCreateOrganization = (name: string) =>
  api.post<ApiResponse<Organization>>('/super-admin/organizations', { name }).then((r) => r.data.data)

// Projects
export const getProjects = (orgId?: string) =>
  api.get<ListResponse<Project>>('/projects', { params: orgId ? { org_id: orgId } : {} }).then((r) => r.data.data)

export const getProject = (id: string) =>
  api.get<ApiResponse<Project>>(`/projects/${id}`).then((r) => r.data.data)

export const createProject = (data: { key: string; name: string; description?: string; owner_id: string; organization_id?: string }) =>
  api.post<ApiResponse<Project>>('/projects', data).then((r) => r.data.data)

export const updateProject = (id: string, data: { name?: string; description?: string }) =>
  api.put<ApiResponse<Project>>(`/projects/${id}`, data).then((r) => r.data.data)

export const deleteProject = (id: string) =>
  api.delete(`/projects/${id}`)

// Statuses
export const getStatuses = (projectId: string) =>
  api.get<ListResponse<Status>>(`/projects/${projectId}/statuses`).then((r) => r.data.data)

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
    workflow_id?: number
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
  workflow_id?: number
}) => api.post<ApiResponse<IssueTemplate>>('/templates', data).then((r) => r.data.data)

export const updateTemplate = (id: number, data: {
  name: string
  description?: string
  body?: string
  default_priority?: string
  workflow_id?: number | null
}) => api.put<ApiResponse<IssueTemplate>>(`/templates/${id}`, data).then((r) => r.data.data)

export const deleteTemplate = (id: number) =>
  api.delete(`/templates/${id}`)

export const updateIssue = (
  projectId: string,
  number: number,
  data: Partial<{ title: string; description: string; status_id: string; priority: string; assignee_id: string }>
) => api.put<ApiResponse<Issue>>(`/projects/${projectId}/issues/${number}`, data).then((r) => r.data.data)

export const deleteIssue = (projectId: string, number: number) =>
  api.delete(`/projects/${projectId}/issues/${number}`)

// Approvals
export const getApprovals = (issueId: string) =>
  api.get<{ data: IssueApproval[] }>(`/issues/${issueId}/approvals`).then((r) => r.data.data)

export const approveStep = (approvalId: string, approverId: string, comment: string) =>
  api.post<ApiResponse<IssueApproval>>(`/approvals/${approvalId}/approve`, { approver_id: approverId, comment }).then((r) => r.data.data)

export const rejectStep = (approvalId: string, approverId: string, comment: string) =>
  api.post<ApiResponse<IssueApproval>>(`/approvals/${approvalId}/reject`, { approver_id: approverId, comment }).then((r) => r.data.data)

// Comments
export const getComments = (issueId: string) =>
  api.get<ListResponse<Comment>>(`/issues/${issueId}/comments`).then((r) => r.data.data)

export const createComment = (issueId: string, data: { author_id: string; body: string }) =>
  api.post<ApiResponse<Comment>>(`/issues/${issueId}/comments`, data).then((r) => r.data.data)

export const deleteComment = (issueId: string, commentId: string) =>
  api.delete(`/issues/${issueId}/comments/${commentId}`)
