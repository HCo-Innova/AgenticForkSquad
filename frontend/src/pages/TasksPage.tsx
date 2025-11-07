import { Link } from 'react-router-dom'
import { useTasks } from '../hooks/useTasks'
import type { Task } from '../services/tasks'

export default function TasksPage() {
  const { data, isLoading, error } = useTasks({ limit: 10, offset: 0 })

  if (isLoading) return <div>Cargando tareas...</div>
  if (error) return <div>Error cargando tareas</div>

  return (
    <div>
      <h2>Tasks</h2>
      <ul>
        {(data?.data ?? []).map((t: Task) => (
          <li key={t.id}>
            <Link to={`/tasks/${t.id}`}>#{t.id} · {t.type} · {t.status}</Link>
          </li>
        ))}
      </ul>
    </div>
  )
}
