import { apiGet } from './api'

export type Agent = {
  type: string
  status?: string
}

export type AgentExecution = {
  id: number
  task_id: number
  agent_type: string
  fork_id: string
  status: string
  started_at: string
  completed_at?: string | null
  error?: string
}

export async function listAgents() {
  return apiGet<{ message?: string; data?: Agent[] }>(`/agents/`)
}

export async function getAgentsByTask(taskId: number) {
  return apiGet<{ data: AgentExecution[] }>(`/tasks/${taskId}/agents`)
}
