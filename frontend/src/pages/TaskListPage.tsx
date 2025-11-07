import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { useTasks } from '../hooks/useTasks'
import type { Task } from '../services/tasks'
import TaskCard from '../components/TaskCard'
import TaskFilters, { type TaskFiltersValue } from '../components/TaskFilters'
import { useWebSocket } from '../hooks/useWebSocket'

export default function TaskListPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  // filtros y paginación básicas
  const [filters, setFilters] = useState<TaskFiltersValue>({})
  const [limit, setLimit] = useState(10)
  const [offset, setOffset] = useState(0)

  const { data, isLoading, error, refetch } = useTasks({
    status: filters.status,
    type: filters.type,
    limit,
    offset,
  })

  // realtime: suscribirse a task_created y refetch
  const ws = useWebSocket(undefined, ['task_created'])
  useEffect(() => {
    const t = setTimeout(() => {
      ws.connect()
    }, 100)
    ws.onMessage((ev) => {
      if (ev.type === 'task_created') {
        // invalidar cache y actualizar
        queryClient.invalidateQueries({ queryKey: ['tasks'] })
      }
    })
    return () => { clearTimeout(t); ws.close() }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function onApplyFilters(v: TaskFiltersValue) {
    setFilters(v)
    setOffset(0)
    refetch()
  }

  function nextPage() {
    if ((data?.pagination?.has_more ?? false)) setOffset(offset + limit)
  }
  function prevPage() {
    setOffset(Math.max(0, offset - limit))
  }

  return (
    <div>
      <h2>Tasks</h2>
      <TaskFilters value={filters} onChange={onApplyFilters} />

      {isLoading && <div>Cargando tareas...</div>}
      {error && <div>Error cargando tareas</div>}

      <div style={{ display: 'grid', gap: 12 }}>
        {(data?.data ?? []).map((t: Task) => (
          <TaskCard key={t.id} task={t} onClick={() => navigate(`/tasks/${t.id}`)} />
        ))}
      </div>

      <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
        <button onClick={prevPage} disabled={offset === 0}>Prev</button>
        <button onClick={nextPage} disabled={!data?.pagination?.has_more}>Next</button>
      </div>
    </div>
  )
}
