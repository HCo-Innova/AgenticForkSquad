import { apiGet } from './api'

export type Proposal = {
  id: number
  agent_execution_id: number
  proposal_type: string
  sql_commands: string[]
  rationale: string
  estimated_impact: Record<string, any>
  created_at: string
}

export async function listProposalsByTask(taskId: number) {
  return apiGet<{ data: Proposal[] }>(`/tasks/${taskId}/proposals`)
}

export type Benchmark = {
  id: number
  proposal_id: number
  query_name: string
  query_executed: string
  execution_time_ms: number
  rows_returned: number
  explain_plan: Record<string, any>
  storage_impact_mb: number
  created_at: string
}

export async function listBenchmarksByProposal(proposalId: number) {
  return apiGet<{ data: Benchmark[] }>(`/proposals/${proposalId}/benchmarks`)
}
