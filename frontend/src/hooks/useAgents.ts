import { useQuery } from '@tanstack/react-query'
import { listAgents, Agent } from '../services/agents'

export function useAgents() {
  return useQuery<{ message?: string; data?: Agent[] }>({
    queryKey: ['agents'],
    queryFn: () => listAgents(),
  })
}
