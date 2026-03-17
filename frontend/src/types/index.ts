export interface User {
  id: string
  name: string
  email: string
  avatar_url?: string
  created_at: string
}

export interface Project {
  id: string
  key: string
  name: string
  description?: string
  owner_id: string
  owner: User
  statuses?: Status[]
  created_at: string
}

export interface Status {
  id: string
  project_id: string
  name: string
  color: string
  order: number
}

export type Priority = 'low' | 'medium' | 'high' | 'critical'

export interface Issue {
  id: string
  number: number
  title: string
  description?: string
  status_id: string
  status: Status
  priority: Priority
  assignee_id?: string
  assignee?: User
  reporter_id: string
  reporter: User
  project_id: string
  due_date?: string
  comments?: Comment[]
  created_at: string
  updated_at: string
}

export interface Comment {
  id: string
  issue_id: string
  author_id: string
  author: User
  body: string
  created_at: string
  updated_at: string
}

export interface ApiResponse<T> {
  data: T
  message?: string
}

export interface ListResponse<T> {
  data: T[]
  total?: number
}

export const PRIORITY_LABELS: Record<Priority, string> = {
  low: '低',
  medium: '中',
  high: '高',
  critical: '緊急',
}

export const PRIORITY_COLORS: Record<Priority, string> = {
  low: 'bg-gray-100 text-gray-700',
  medium: 'bg-blue-100 text-blue-700',
  high: 'bg-orange-100 text-orange-700',
  critical: 'bg-red-100 text-red-700',
}
