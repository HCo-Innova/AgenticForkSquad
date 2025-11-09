import { useQuery } from '@tanstack/react-query'
import { useApi } from './useApi'

interface DashboardMetrics {
  total_tasks: number
  completed_tasks: number
  failed_tasks: number
  in_progress_tasks: number
  success_rate: number
  avg_duration_seconds: number
  total_optimizations: number
  avg_improvement_percent: number
}

interface AgentMetrics {
  name: string
  total_tasks: number
  wins: number
  success_rate: number
  win_rate: number
  avg_duration: number
}

interface PerformanceData {
  date: string
  success_rate: number
  avg_duration: number
  tasks: number
}

export function useMetrics() {
  const { fetchWithAuth } = useApi()

  const overview = useQuery<DashboardMetrics>({
    queryKey: ['metrics', 'overview'],
    queryFn: () => fetchWithAuth('/api/v1/metrics/overview'),
    refetchInterval: 30000, // 30s
  })

  const agents = useQuery<AgentMetrics[]>({
    queryKey: ['metrics', 'agents'],
    queryFn: () => fetchWithAuth('/api/v1/metrics/agents'),
    refetchInterval: 30000,
  })

  const performance = useQuery<PerformanceData[]>({
    queryKey: ['metrics', 'performance'],
    queryFn: () => fetchWithAuth('/api/v1/metrics/performance?days=7'),
    refetchInterval: 60000, // 1min
  })

  return { overview, agents, performance }
}
