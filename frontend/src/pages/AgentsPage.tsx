import { useAgents } from '../hooks/useAgents'

export default function AgentsPage() {
  const { data, isLoading, error } = useAgents()

  if (isLoading) return <div>Cargando agentes...</div>
  if (error) return <div>Error cargando agentes</div>

  const list = data?.data || []

  return (
    <div>
      <h2>Agents</h2>
      {list.length === 0 ? (
        <p>No hay agentes listados (endpoint aún en TODO)</p>
      ) : (
        <ul>
          {list.map((a, idx) => (
            <li key={idx}>{a.type} · {a.status ?? 'unknown'}</li>
          ))}
        </ul>
      )}
    </div>
  )
}
