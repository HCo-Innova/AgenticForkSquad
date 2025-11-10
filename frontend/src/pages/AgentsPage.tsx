import { useAgents } from '../hooks/useAgents'

export default function AgentsPage() {
  const { data, isLoading, error } = useAgents()

  if (isLoading) return <div className="p-8">Loading agents...</div>
  if (error) return <div className="p-8 text-red-600">Error loading agents</div>

  const list = data?.data || []

  return (
    <div className="p-8">
      <h2 className="text-2xl font-bold mb-6">🤖 Agent Executions</h2>
      <p className="text-gray-600 mb-4">History of agent executions in the system</p>
      
      {list.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          <p>No agents executed yet</p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full bg-white border border-gray-200 rounded-lg">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Task</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Agent Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Fork ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Started</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Duration</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {list.map((a: any) => {
                const started = new Date(a.started_at)
                const completed = a.completed_at ? new Date(a.completed_at) : null
                const duration = completed ? Math.round((completed.getTime() - started.getTime()) / 1000) : null
                
                return (
                  <tr key={a.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm">{a.id}</td>
                    <td className="px-4 py-3 text-sm">
                      <a href={`/tasks/${a.task_id}`} className="text-blue-600 hover:underline">
                        #{a.task_id}
                      </a>
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <span className={`px-2 py-1 rounded text-xs font-medium ${
                        a.agent_type === 'cerebro' ? 'bg-purple-100 text-purple-800' :
                        a.agent_type === 'operativo' ? 'bg-blue-100 text-blue-800' :
                        'bg-green-100 text-green-800'
                      }`}>
                        {a.agent_type}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm font-mono text-xs">{a.fork_id}</td>
                    <td className="px-4 py-3 text-sm">
                      <span className={`px-2 py-1 rounded text-xs font-medium ${
                        a.status === 'completed' ? 'bg-green-100 text-green-800' :
                        a.status === 'failed' ? 'bg-red-100 text-red-800' :
                        'bg-yellow-100 text-yellow-800'
                      }`}>
                        {a.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">
                      {started.toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">
                      {duration ? `${duration}s` : '—'}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
