import axios from 'axios'
import type { ApiResponse, ListResponse, Project, Issue, Comment, User, Status } from '@/types'

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

// Users
export const getUsers = () =>
  api.get<ListResponse<User>>('/users').then((r) => r.data.data)

export const createUser = (data: { name: string; email: string }) =>
  api.post<ApiResponse<User>>('/users', data).then((r) => r.data.data)

// Projects
export const getProjects = () =>
  api.get<ListResponse<Project>>('/projects').then((r) => r.data.data)

export const getProject = (id: string) =>
  api.get<ApiResponse<Project>>(`/projects/${id}`).then((r) => r.data.data)

export const createProject = (data: { key: string; name: string; description?: string; owner_id: string }) =>
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
  }
) => api.post<ApiResponse<Issue>>(`/projects/${projectId}/issues`, data).then((r) => r.data.data)

export const updateIssue = (
  projectId: string,
  number: number,
  data: Partial<{ title: string; description: string; status_id: string; priority: string; assignee_id: string }>
) => api.put<ApiResponse<Issue>>(`/projects/${projectId}/issues/${number}`, data).then((r) => r.data.data)

export const deleteIssue = (projectId: string, number: number) =>
  api.delete(`/projects/${projectId}/issues/${number}`)

// Comments
export const getComments = (issueId: string) =>
  api.get<ListResponse<Comment>>(`/issues/${issueId}/comments`).then((r) => r.data.data)

export const createComment = (issueId: string, data: { author_id: string; body: string }) =>
  api.post<ApiResponse<Comment>>(`/issues/${issueId}/comments`, data).then((r) => r.data.data)

export const deleteComment = (issueId: string, commentId: string) =>
  api.delete(`/issues/${issueId}/comments/${commentId}`)
