import { useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { listProposalsByTask, listBenchmarksByProposal, type Proposal, type Benchmark } from '../../services/optimizations'
import ComparisonTable from './ComparisonTable'
import BenchmarksChart from './BenchmarksChart'
import ScoreBreakdown from './ScoreBreakdown'

export default function OptimizationDashboard({ taskId }: { taskId: number }) {
  const qc = useQueryClient()
  const [selectedId, setSelectedId] = useState<number | undefined>(undefined)

  const proposalsQ = useQuery<{ data: Proposal[] }>({
    queryKey: ['proposals', taskId],
    queryFn: () => listProposalsByTask(taskId),
    enabled: !!taskId,
  })

  const selectedProposal = useMemo(
    () => proposalsQ.data?.data.find((p) => p.id === selectedId) ?? proposalsQ.data?.data[0],
    [proposalsQ.data, selectedId]
  )

  useEffect(() => {
    if (!selectedId && proposalsQ.data?.data?.length) setSelectedId(proposalsQ.data.data[0].id)
  }, [selectedId, proposalsQ.data])

  const benchmarksQ = useQuery<{ data: Benchmark[] }>({
    queryKey: ['benchmarks', selectedProposal?.id],
    queryFn: () => listBenchmarksByProposal(selectedProposal!.id),
    enabled: !!selectedProposal?.id,
  })

  return (
    <div style={{ display: 'grid', gap: 16 }}>
      <ComparisonTable proposals={proposalsQ.data?.data ?? []} onSelect={setSelectedId} selectedId={selectedId} />

      {selectedProposal ? (
        <div style={{ display: 'grid', gap: 16 }}>
          <ScoreBreakdown proposal={selectedProposal} />
          <BenchmarksChart items={benchmarksQ.data?.data ?? []} />
        </div>
      ) : (
        <div>Sin propuestas</div>
      )}
    </div>
  )
}
