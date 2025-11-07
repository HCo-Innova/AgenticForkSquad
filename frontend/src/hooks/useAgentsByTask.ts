import { useQuery } from '@tanstack/react-query'
import { getAgentsByTask, type AgentExecution } from '../services/agents'

export function useAgentsByTask(taskId: number) {
  return useQuery<{ data: AgentExecution[] }>({
    queryKey: ['agentsByTask', taskId],
    queryFn: () => getAgentsByTask(taskId),
    enabled: !!taskId,
  })
}
