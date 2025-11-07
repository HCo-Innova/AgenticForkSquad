import { useEffect, useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useTask } from '../hooks/useTasks'
import { useAgentsByTask } from '../hooks/useAgentsByTask'
import { useQueryClient } from '@tanstack/react-query'
import { useWebSocket, type WSEvent } from '../hooks/useWebSocket'
import OptimizationDashboard from '../components/optimization/OptimizationDashboard'

type TimelineItem = { ts: string; type: string; payload?: Record<string, any> }

export default function TaskDetailPage() {
  const { id } = useParams()
  const taskId = Number(id)
  const { data, isLoading, error } = useTask(taskId)
  const agentsQ = useAgentsByTask(taskId)
  const qc = useQueryClient()

  const [timeline, setTimeline] = useState<TimelineItem[]>([])

  const ws = useWebSocket(undefined, [
    'task_created',
    'agents_assigned',
    'fork_created',
    'analysis_completed',
    'proposal_submitted',
    'benchmark_completed',
    'consensus_reached',
    'optimization_applied',
    'task_completed',
    'task_failed',
  ])

  useEffect(() => {
    if (!taskId) return
    ws.connect()
    ws.onMessage((ev: WSEvent) => {
      // filtrar por task
      const pid = (ev.payload?.task_id as number | undefined) ?? (ev.payload?.id as number | undefined)
      if (pid !== taskId) return
      setTimeline((t) => [
        { ts: new Date().toISOString(), type: ev.type, payload: ev.payload },
        ...t,
      ])
      // refrescar datos relevantes
      switch (ev.type) {
        // Do NOT invalidate on 'agents_assigned' to avoid loop when backend broadcasts on GET /agents
        case 'analysis_completed':
        case 'task_completed':
        case 'task_failed':
          qc.invalidateQueries({ queryKey: ['task', taskId] })
          qc.invalidateQueries({ queryKey: ['agentsByTask', taskId] })
          break
      }
    })
    return () => ws.close()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [taskId])

  const agents = useMemo(() => agentsQ.data?.data ?? [], [agentsQ.data])

  if (!id || Number.isNaN(taskId)) return <div>ID inválido</div>
  if (isLoading) return <div>Cargando tarea...</div>
  if (error) return <div>Error cargando tarea</div>

  return (
    <div style={{ display: 'grid', gap: 16 }}>
      <div>
        <h2>Task #{data?.id}</h2>
        <p>Type: {data?.type}</p>
        <p>Status: {data?.status}</p>
        <p>Target Query: <code>{data?.target_query}</code></p>
        <p>Created: {data?.created_at}</p>
      </div>

      <div>
        <h3>Agents</h3>
        {agentsQ.isLoading && <div>Cargando agentes...</div>}
        {agentsQ.error && <div>Error cargando agentes</div>}
        {agents.length === 0 ? (
          <div>No hay ejecuciones de agentes</div>
        ) : (
          <ul style={{ display: 'grid', gap: 8, listStyle: 'none', padding: 0 }}>
            {agents.map((a) => (
              <li key={a.id} style={{ border: '1px solid #eee', borderRadius: 8, padding: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <strong>{a.agent_type}</strong>
                  <span>{a.status}</span>
                </div>
                <div style={{ color: '#6b7280', fontSize: 12 }}>Fork: {a.fork_id}</div>
                <div style={{ color: '#6b7280', fontSize: 12 }}>Started: {a.started_at}</div>
                {a.completed_at && <div style={{ color: '#6b7280', fontSize: 12 }}>Completed: {a.completed_at}</div>}
                {a.error && <div style={{ color: 'crimson', fontSize: 12 }}>Error: {a.error}</div>}
              </li>
            ))}
          </ul>
        )}
      </div>

      <div>
        <h3>Timeline</h3>
        {timeline.length === 0 ? (
          <div>Sin eventos aún</div>
        ) : (
          <ul style={{ listStyle: 'none', padding: 0, display: 'grid', gap: 8 }}>
            {timeline.map((it, idx) => (
              <li key={idx} style={{ border: '1px dashed #ddd', borderRadius: 8, padding: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <strong>{it.type}</strong>
                  <span style={{ color: '#6b7280', fontSize: 12 }}>{it.ts}</span>
                </div>
                {it.payload && (
                  <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{JSON.stringify(it.payload, null, 2)}</pre>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      <div>
        <h3>Optimization Dashboard</h3>
        <OptimizationDashboard taskId={taskId} />
      </div>
    </div>
  )
}
