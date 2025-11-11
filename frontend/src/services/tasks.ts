import { apiGet, apiPost, apiDelete } from './api'

export type Task = {
  id: number
  type: string
  description?: string
  target_query: string
  status: string
  created_at: string
  completed_at?: string
}

export type Paginated<T> = {
  data: T[]
  pagination: { limit: number; offset: number; total: number; has_more: boolean }
}

export async function listTasks(params?: { status?: string; type?: string; limit?: number; offset?: number }) {
  const q = new URLSearchParams()
  if (params?.status) q.set('status', params.status)
  if (params?.type) q.set('type', params.type)
  if (params?.limit != null) q.set('limit', String(params.limit))
  if (params?.offset != null) q.set('offset', String(params.offset))
  const qs = q.toString() ? `?${q.toString()}` : ''
  return apiGet<Paginated<Task>>(`/api/v1/tasks${qs}`)
}

export async function getTask(id: number) {
  return apiGet<Task>(`/api/v1/tasks/${id}`)
}

export type CreateTaskInput = {
  type: string
  description?: string
  target_query: string
}

export async function createTask(input: CreateTaskInput) {
  return apiPost<Task>(`/api/v1/tasks`, input)
}

export async function deleteTask(id: number) {
  return apiDelete(`/api/v1/tasks/${id}`)
}
