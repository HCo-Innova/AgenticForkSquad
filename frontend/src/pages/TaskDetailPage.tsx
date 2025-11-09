import { useEffect, useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useTask } from '../hooks/useTasks'
import { useAgentsByTask } from '../hooks/useAgentsByTask'
import { useQueryClient } from '@tanstack/react-query'
import { useWebSocket, type WSEvent } from '../hooks/useWebSocket'
import OptimizationDashboard from '../components/optimization/OptimizationDashboard'
import SQLViewer from '../components/SQLViewer'

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

  // Helper para mapear status a colores
  const getStatusColor = (status: string) => {
    const colors = {
      pending: 'bg-yellow-100 text-yellow-800 border-yellow-200',
      routing: 'bg-blue-100 text-blue-800 border-blue-200',
      executing: 'bg-purple-100 text-purple-800 border-purple-200',
      consensus: 'bg-indigo-100 text-indigo-800 border-indigo-200',
      completed: 'bg-green-100 text-green-800 border-green-200',
      failed: 'bg-red-100 text-red-800 border-red-200',
    }
    return colors[status as keyof typeof colors] || 'bg-gray-100 text-gray-800 border-gray-200'
  }

  const getAgentIcon = (type: string) => {
    const icons = {
      planner: '🧠',
      generator: '⚡',
      operator: '🔧',
    }
    return icons[type as keyof typeof icons] || '🤖'
  }

  if (!id || Number.isNaN(taskId)) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-md">
          ❌ ID de tarea inválido
        </div>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          <span className="ml-3 text-gray-600">Cargando tarea...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-md">
          ❌ Error cargando tarea
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-7xl mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="bg-white shadow-md rounded-lg p-6">
        <div className="flex items-start justify-between mb-4">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Task #{data?.id}</h1>
            <p className="text-gray-600 mt-1">{data?.description || 'Sin descripción'}</p>
          </div>
          <span className={`px-4 py-2 rounded-full text-sm font-semibold border ${getStatusColor(data?.status || 'pending')}`}>
            {data?.status?.toUpperCase()}
          </span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
          <div className="bg-gray-50 p-4 rounded-lg">
            <div className="text-sm text-gray-500 font-medium">Tipo</div>
            <div className="text-lg font-semibold text-gray-900 mt-1">
              {data?.type?.replace('_', ' ').toUpperCase()}
            </div>
          </div>
          <div className="bg-gray-50 p-4 rounded-lg">
            <div className="text-sm text-gray-500 font-medium">Creado</div>
            <div className="text-lg font-semibold text-gray-900 mt-1">
              {data?.created_at ? new Date(data.created_at).toLocaleString('es-ES') : 'N/A'}
            </div>
          </div>
          <div className="bg-gray-50 p-4 rounded-lg">
            <div className="text-sm text-gray-500 font-medium">Agentes Asignados</div>
            <div className="text-lg font-semibold text-gray-900 mt-1">
              {agents.length} / 3
            </div>
          </div>
        </div>

        <div className="mt-6">
          <h4 className="text-sm font-medium text-gray-700 mb-2">🎯 Query a Optimizar</h4>
          <SQLViewer code={data?.target_query || ''} height="200px" />
        </div>
      </div>

      {/* Agents Section */}
      <div className="bg-white shadow-md rounded-lg p-6">
        <h3 className="text-2xl font-bold text-gray-900 mb-4">🤖 Agentes Multi-Especializados</h3>
        {agentsQ.isLoading && (
          <div className="flex items-center text-gray-600">
            <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600 mr-2"></div>
            Cargando agentes...
          </div>
        )}
        {agentsQ.error && (
          <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-md">
            Error cargando agentes
          </div>
        )}
        {agents.length === 0 && !agentsQ.isLoading ? (
          <div className="bg-yellow-50 border border-yellow-200 text-yellow-800 px-4 py-3 rounded-md flex items-center">
            <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            No hay ejecuciones de agentes aún. El sistema asignará 3 agentes especializados para analizar el query.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {agents.map((a) => (
              <div key={a.id} className="border border-gray-200 rounded-lg p-4 hover:shadow-lg transition-shadow">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center">
                    <span className="text-2xl mr-2">{getAgentIcon(a.agent_type)}</span>
                    <strong className="text-lg text-gray-900">{a.agent_type}</strong>
                  </div>
                  <span className={`px-2 py-1 rounded text-xs font-semibold ${getStatusColor(a.status)}`}>
                    {a.status}
                  </span>
                </div>
                <div className="space-y-1 text-sm text-gray-600">
                  <div className="flex items-center">
                    <span className="font-medium mr-2">🍴 Fork:</span>
                    <code className="bg-gray-100 px-2 py-0.5 rounded text-xs">{a.fork_id || 'N/A'}</code>
                  </div>
                  <div className="flex items-center">
                    <span className="font-medium mr-2">⏱️ Inicio:</span>
                    <span>{a.started_at ? new Date(a.started_at).toLocaleTimeString('es-ES') : 'N/A'}</span>
                  </div>
                  {a.completed_at && (
                    <div className="flex items-center text-green-600">
                      <span className="font-medium mr-2">✅ Fin:</span>
                      <span>{new Date(a.completed_at).toLocaleTimeString('es-ES')}</span>
                    </div>
                  )}
                  {a.error && (
                    <div className="bg-red-50 border border-red-200 text-red-800 px-2 py-1 rounded text-xs mt-2">
                      ❌ {a.error}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Timeline Section */}
      <div className="bg-white shadow-md rounded-lg p-6">
        <h3 className="text-2xl font-bold text-gray-900 mb-4">📊 Timeline de Eventos en Tiempo Real</h3>
        {timeline.length === 0 ? (
          <div className="bg-gray-50 border border-gray-200 text-gray-600 px-4 py-8 rounded-md text-center">
            <svg className="w-12 h-12 mx-auto mb-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p>Sin eventos aún. Los eventos aparecerán aquí en tiempo real vía WebSocket.</p>
          </div>
        ) : (
          <div className="space-y-3 max-h-96 overflow-y-auto">
            {timeline.map((it, idx) => (
              <div key={idx} className="border-l-4 border-blue-500 bg-blue-50 pl-4 pr-4 py-3 rounded-r-lg">
                <div className="flex items-center justify-between mb-2">
                  <strong className="text-blue-900 font-semibold">{it.type.replace(/_/g, ' ').toUpperCase()}</strong>
                  <span className="text-xs text-gray-500">
                    {new Date(it.ts).toLocaleTimeString('es-ES', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                  </span>
                </div>
                {it.payload && (
                  <details className="mt-2">
                    <summary className="cursor-pointer text-xs text-blue-700 hover:text-blue-900">Ver payload</summary>
                    <pre className="mt-2 text-xs bg-white border border-blue-200 rounded p-2 overflow-x-auto">
                      {JSON.stringify(it.payload, null, 2)}
                    </pre>
                  </details>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Optimization Dashboard */}
      <div className="bg-white shadow-md rounded-lg p-6">
        <h3 className="text-2xl font-bold text-gray-900 mb-4">⚖️ Comparison Dashboard</h3>
        <OptimizationDashboard taskId={taskId} />
      </div>
    </div>
  )
}
