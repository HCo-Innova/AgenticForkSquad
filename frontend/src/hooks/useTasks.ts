import { useQuery } from '@tanstack/react-query'
import { getTask, listTasks, Task, Paginated } from '../services/tasks'

export function useTasks(params?: { status?: string; type?: string; limit?: number; offset?: number }) {
  return useQuery<Paginated<Task>>({
    queryKey: ['tasks', params],
    queryFn: () => listTasks(params),
  })
}

export function useTask(id: number) {
  return useQuery<Task>({
    queryKey: ['task', id],
    queryFn: () => getTask(id),
    enabled: !!id,
  })
}
