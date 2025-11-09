import { useMetrics } from '../hooks/useMetrics'
import { LineChart, Line, BarChart, Bar, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'

export default function AnalyticsPage() {
  const { overview, agents, performance } = useMetrics()

  const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']

  const agentDistribution = agents.data?.map((agent: any, idx: number) => ({
    name: agent.name,
    tasks: agent.total_tasks,
    color: COLORS[idx % COLORS.length]
  })) || []

  const successRateData = agents.data?.map((agent: any) => ({
    name: agent.name,
    'Success Rate': agent.success_rate,
    'Win Rate': agent.win_rate
  })) || []

  return (
    <div className="max-w-7xl mx-auto py-6 space-y-8">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-900">Analytics Dashboard</h1>
        <div className="text-sm text-gray-500">
          Auto-refreshing every 30 seconds
        </div>
      </div>

      {/* Performance Trend */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">Performance Trend (7 Days)</h2>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={performance.data || []}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="success_rate" stroke="#10b981" strokeWidth={2} name="Success Rate %" />
            <Line type="monotone" dataKey="avg_duration" stroke="#3b82f6" strokeWidth={2} name="Avg Duration (s)" />
            <Line type="monotone" dataKey="tasks" stroke="#f59e0b" strokeWidth={2} name="Tasks Completed" />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Agent Success & Win Rates */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Agent Performance Comparison</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={successRateData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis domain={[0, 100]} />
              <Tooltip />
              <Legend />
              <Bar dataKey="Success Rate" fill="#10b981" />
              <Bar dataKey="Win Rate" fill="#3b82f6" />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Task Distribution by Agent */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Task Distribution by Agent</h2>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={agentDistribution}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={(entry) => `${entry.name}: ${entry.tasks}`}
                outerRadius={100}
                fill="#8884d8"
                dataKey="tasks"
              >
                {agentDistribution.map((entry: any, index: number) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Detailed Statistics Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Detailed Agent Statistics</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Agent</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Total Tasks</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Wins</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Success Rate</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Win Rate</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Avg Duration</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {agents.data?.map((agent: any) => (
                <tr key={agent.name} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-gray-900">{agent.name}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">{agent.total_tasks}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">{agent.wins}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      agent.success_rate >= 80 ? 'bg-green-100 text-green-800' :
                      agent.success_rate >= 60 ? 'bg-yellow-100 text-yellow-800' :
                      'bg-red-100 text-red-800'
                    }`}>
                      {agent.success_rate.toFixed(1)}%
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      agent.win_rate >= 40 ? 'bg-blue-100 text-blue-800' :
                      agent.win_rate >= 20 ? 'bg-gray-100 text-gray-800' :
                      'bg-red-100 text-red-800'
                    }`}>
                      {agent.win_rate.toFixed(1)}%
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">{agent.avg_duration.toFixed(2)}s</div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg shadow p-6 text-white">
          <div className="text-sm font-medium opacity-90">Total Tasks</div>
          <div className="text-3xl font-bold mt-2">{overview.data?.total_tasks || 0}</div>
        </div>
        <div className="bg-gradient-to-br from-green-500 to-green-600 rounded-lg shadow p-6 text-white">
          <div className="text-sm font-medium opacity-90">Success Rate</div>
          <div className="text-3xl font-bold mt-2">{overview.data?.success_rate.toFixed(1) || 0}%</div>
        </div>
        <div className="bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg shadow p-6 text-white">
          <div className="text-sm font-medium opacity-90">Avg Duration</div>
          <div className="text-3xl font-bold mt-2">{overview.data?.avg_duration_seconds.toFixed(1) || 0}s</div>
        </div>
        <div className="bg-gradient-to-br from-orange-500 to-orange-600 rounded-lg shadow p-6 text-white">
          <div className="text-sm font-medium opacity-90">Optimizations</div>
          <div className="text-3xl font-bold mt-2">{overview.data?.total_optimizations || 0}</div>
        </div>
      </div>
    </div>
  )
}
