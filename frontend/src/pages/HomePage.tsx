import { useMetrics } from '../hooks/useMetrics'
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { motion } from 'framer-motion'

export default function HomePage() {
  const { overview, agents, performance } = useMetrics()

  const formatDuration = (seconds: number) => {
    if (seconds < 60) return `${Math.round(seconds)}s`
    const mins = Math.floor(seconds / 60)
    const secs = Math.round(seconds % 60)
    return `${mins}m ${secs}s`
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-4xl font-bold text-gray-900 mb-2">
          🚀 Agentic Fork Squad
        </h1>
        <p className="text-gray-600">Multi-Agent Database Optimization System</p>
      </div>

      {/* Metrics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <MetricCard
          title="Total Tasks"
          value={overview.data?.total_tasks || 0}
          icon="📊"
          color="blue"
        />
        <MetricCard
          title="Success Rate"
          value={`${(overview.data?.success_rate || 0).toFixed(1)}%`}
          icon="✅"
          color="green"
        />
        <MetricCard
          title="Avg Duration"
          value={formatDuration(overview.data?.avg_duration_seconds || 0)}
          icon="⏱️"
          color="purple"
        />
        <MetricCard
          title="Optimizations"
          value={overview.data?.total_optimizations || 0}
          icon="🎯"
          color="orange"
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        {/* Performance Chart */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Performance Trend (7 days)</h2>
          {performance.isLoading ? (
            <div className="h-64 flex items-center justify-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
            </div>
          ) : (
            <ResponsiveContainer width="100%" height={250}>
              <LineChart data={performance.data || []}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line type="monotone" dataKey="tasks" stroke="#3B82F6" name="Tasks" />
                <Line type="monotone" dataKey="success_rate" stroke="#10B981" name="Improvement %" />
              </LineChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* Agent Performance */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Agent Performance</h2>
          {agents.isLoading ? (
            <div className="h-64 flex items-center justify-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
            </div>
          ) : (
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={agents.data || []}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="agent_type" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar dataKey="win_rate" fill="#3B82F6" name="Win Rate %" />
                <Bar dataKey="success_rate" fill="#10B981" name="Success Rate %" />
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>

      {/* Agent Stats Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold">Agent Statistics</h2>
        </div>
        <div className="overflow-x-auto">
          {agents.isLoading ? (
            <div className="h-48 flex items-center justify-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
            </div>
          ) : (
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Agent</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Total Tasks</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Success Rate</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Win Rate</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Avg Duration</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {agents.data?.map((agent) => (
                  <tr key={agent.agent_type} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap font-medium text-gray-900">
                      {agent.agent_type}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-gray-600">
                      {agent.total_tasks}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 py-1 rounded-full text-xs font-semibold ${
                        agent.success_rate >= 80 ? 'bg-green-100 text-green-800' :
                        agent.success_rate >= 50 ? 'bg-yellow-100 text-yellow-800' :
                        'bg-red-100 text-red-800'
                      }`}>
                        {agent.success_rate.toFixed(1)}%
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-blue-600 font-semibold">
                        {agent.win_rate.toFixed(1)}%
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-gray-600">
                      {formatDuration(agent.avg_duration || 0)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* System Health */}
      <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-6">
        <HealthCard title="Tiger Cloud" status="operational" />
        <HealthCard title="Vertex AI" status="operational" />
        <HealthCard title="Database" status="operational" />
      </div>
    </div>
  )
}

function MetricCard({ title, value, icon, color }: { title: string; value: string | number; icon: string; color: string }) {
  const colors = {
    blue: 'from-blue-500 to-blue-600',
    green: 'from-green-500 to-green-600',
    purple: 'from-purple-500 to-purple-600',
    orange: 'from-orange-500 to-orange-600',
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className={`bg-gradient-to-br ${colors[color as keyof typeof colors]} rounded-lg shadow-lg p-6 text-white`}
    >
      <div className="flex items-center justify-between mb-2">
        <span className="text-3xl">{icon}</span>
        <div className="text-right">
          <p className="text-sm opacity-90">{title}</p>
          <p className="text-3xl font-bold">{value}</p>
        </div>
      </div>
    </motion.div>
  )
}

function HealthCard({ title, status }: { title: string; status: 'operational' | 'degraded' | 'down' }) {
  const statusColors = {
    operational: 'bg-green-100 text-green-800',
    degraded: 'bg-yellow-100 text-yellow-800',
    down: 'bg-red-100 text-red-800',
  }

  const statusDots = {
    operational: 'bg-green-500',
    degraded: 'bg-yellow-500',
    down: 'bg-red-500',
  }

  return (
    <div className="bg-white rounded-lg shadow p-4 flex items-center justify-between">
      <div className="flex items-center">
        <div className={`w-3 h-3 rounded-full ${statusDots[status]} mr-3`}></div>
        <span className="font-medium text-gray-900">{title}</span>
      </div>
      <span className={`px-2 py-1 rounded-full text-xs font-semibold ${statusColors[status]}`}>
        {status}
      </span>
    </div>
  )
}
