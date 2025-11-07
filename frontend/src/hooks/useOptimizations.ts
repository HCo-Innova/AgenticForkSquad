import { useQuery } from '@tanstack/react-query'
import { listProposalsByTask, listBenchmarksByProposal, Proposal, Benchmark } from '../services/optimizations'

export function useProposals(taskId: number) {
  return useQuery<{ data: Proposal[] }>({
    queryKey: ['proposals', taskId],
    queryFn: () => listProposalsByTask(taskId),
    enabled: !!taskId,
  })
}

export function useBenchmarks(proposalId: number) {
  return useQuery<{ data: Benchmark[] }>({
    queryKey: ['benchmarks', proposalId],
    queryFn: () => listBenchmarksByProposal(proposalId),
    enabled: !!proposalId,
  })
}
