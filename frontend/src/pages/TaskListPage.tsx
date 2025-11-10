import { Link } from 'react-router-dom'
import { useState } from 'react'
import { useTasks } from '../hooks/useTasks'
import { deleteTask } from '../services/tasks'
import type { Task } from '../services/tasks'
import { useQueryClient } from '@tanstack/react-query'

export default function TaskListPage() {
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [typeFilter, setTypeFilter] = useState<string>('all')
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<number | null>(null)
  const [offset, setOffset] = useState(0)
  const limit = 10
  const queryClient = useQueryClient()
  
  const { data, isLoading, error } = useTasks({ limit: 50, offset: 0 })

  const handleDelete = async (taskId: number) => {
    try {
      await deleteTask(taskId)
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      setShowDeleteConfirm(null)
    } catch (err) {
      console.error('Failed to delete task:', err)
      alert('Failed to delete task. Please try again.')
    }
  }

  if (isLoading) return (
    <div className="flex items-center justify-center h-64">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
    </div>
  )
  
  if (error) return (
    <div className="p-8 text-center text-red-600">
      <p className="text-lg font-semibold">Error cargando tareas</p>
    </div>
  )

  const tasks = data?.data ?? []
  const filteredTasks = tasks.filter((t: Task) => {
    const statusMatch = statusFilter === 'all' || t.status === statusFilter
    const typeMatch = typeFilter === 'all' || t.type === typeFilter
    return statusMatch && typeMatch
  })

  const paginatedTasks = filteredTasks.slice(offset, offset + limit)
  const totalPages = Math.ceil(filteredTasks.length / limit)
  const currentPage = Math.floor(offset / limit) + 1

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      completed: 'bg-green-100 text-green-800 border-green-200',
      failed: 'bg-red-100 text-red-800 border-red-200',
      in_progress: 'bg-blue-100 text-blue-800 border-blue-200',
      pending: 'bg-yellow-100 text-yellow-800 border-yellow-200'
    }
    return colors[status] || 'bg-gray-100 text-gray-800 border-gray-200'
  }

  const getStatusIcon = (status: string) => {
    const icons: Record<string, string> = {
      completed: '✅',
      failed: '❌',
      in_progress: '⏳',
      pending: '⏸️'
    }
    return icons[status] || '📋'
  }

  const truncateQuery = (query: string, maxLength: number = 100) => {
    if (query.length <= maxLength) return query
    return query.substring(0, maxLength) + '...'
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">📋 Tasks</h1>
        <p className="text-gray-600">Manage and monitor optimization tasks</p>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <div className="flex flex-wrap items-center gap-6">
          <div className="flex items-center gap-3">
            <label className="text-sm font-semibold text-gray-700 min-w-[60px]">Status</label>
            <select
              value={statusFilter}
              onChange={(e) => { setStatusFilter(e.target.value); setOffset(0) }}
              className="px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white text-gray-700 font-medium min-w-[150px]"
            >
              <option value="all">All Statuses</option>
              <option value="completed">Completed</option>
              <option value="failed">Failed</option>
              <option value="in_progress">In Progress</option>
              <option value="pending">Pending</option>
            </select>
          </div>

          <div className="flex items-center gap-3">
            <label className="text-sm font-semibold text-gray-700 min-w-[60px]">Type</label>
            <select
              value={typeFilter}
              onChange={(e) => { setTypeFilter(e.target.value); setOffset(0) }}
              className="px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white text-gray-700 font-medium min-w-[180px]"
            >
              <option value="all">All Types</option>
              <option value="query_optimization">Query Optimization</option>
              <option value="schema_review">Schema Review</option>
            </select>
          </div>

          <div className="ml-auto flex items-center gap-2 text-sm text-gray-600 bg-gray-50 px-4 py-2 rounded-lg">
            <span className="font-semibold">{filteredTasks.length}</span>
            <span>of</span>
            <span className="font-semibold">{tasks.length}</span>
            <span>tasks</span>
          </div>
        </div>
      </div>

      {/* Tasks List */}
      <div className="space-y-3">
        {paginatedTasks.map((task: Task) => (
          <div key={task.id} className="bg-white rounded-lg shadow-sm border border-gray-200 hover:shadow-md hover:border-blue-300 transition-all">
            <Link to={`/tasks/${task.id}`} className="block p-5 cursor-pointer">
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <span className="text-xl font-bold text-blue-600 hover:text-blue-800">
                    #{task.id}
                  </span>
                  <span className={`px-3 py-1.5 rounded-full text-xs font-semibold border ${getStatusColor(task.status)}`}>
                    {getStatusIcon(task.status)} {task.status.replace('_', ' ').toUpperCase()}
                  </span>
                </div>
                
                <button
                  onClick={(e) => { e.preventDefault(); e.stopPropagation(); setShowDeleteConfirm(task.id); }}
                  className="flex items-center gap-2 text-red-600 hover:text-white hover:bg-red-600 text-sm font-semibold px-4 py-2 rounded-lg border border-red-300 hover:border-red-600 transition-colors"
                >
                  <span>🗑️</span>
                  <span>Delete</span>
                </button>
              </div>

              <div className="mb-3">
                <span className="text-sm font-semibold text-gray-500">Type:</span>{' '}
                <span className="text-sm text-gray-900 font-medium">{task.type}</span>
              </div>

              {task.target_query && (
                <div className="mb-3">
                  <span className="text-sm font-semibold text-gray-500 block mb-1">Query:</span>
                  <div className="bg-gray-50 rounded-lg p-3 border border-gray-200">
                    <code className="text-xs text-gray-700 font-mono block">
                      {truncateQuery(task.target_query, 100)}
                    </code>
                    {task.target_query.length > 100 && (
                      <span className="inline-block mt-2 text-xs text-blue-600 hover:text-blue-800 font-semibold">
                        → View full query
                      </span>
                    )}
                  </div>
                </div>
              )}

              <div className="flex items-center gap-4 text-xs text-gray-500">
                <span className="flex items-center gap-1">
                  <span className="font-semibold">Created:</span>
                  <span>{new Date(task.created_at).toLocaleString()}</span>
                </span>
                {task.completed_at && (
                  <span className="flex items-center gap-1">
                    <span className="font-semibold">Completed:</span>
                    <span>{new Date(task.completed_at).toLocaleString()}</span>
                  </span>
                )}
              </div>
            </Link>

            {/* Delete Confirmation */}
            {showDeleteConfirm === task.id && (
              <div className="border-t border-red-200 bg-red-50 px-5 py-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-semibold text-gray-900 mb-1">⚠️ Confirm Deletion</p>
                    <p className="text-sm text-gray-700">
                      Delete task #{task.id}? This action cannot be undone.
                    </p>
                  </div>
                  <div className="flex gap-3">
                    <button
                      onClick={() => setShowDeleteConfirm(null)}
                      className="px-4 py-2 text-sm font-semibold text-gray-700 bg-white border-2 border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={() => handleDelete(task.id)}
                      className="px-4 py-2 text-sm font-semibold text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors shadow-sm"
                    >
                      Delete Task
                    </button>
                  </div>
                </div>
              </div>
            )}
          </div>
        ))}

        {paginatedTasks.length === 0 && (
          <div className="text-center py-16 bg-white rounded-lg shadow-sm border border-gray-200">
            <div className="text-6xl mb-4">📭</div>
            <p className="text-lg font-semibold text-gray-700 mb-2">No tasks found</p>
            <p className="text-gray-500">Try adjusting your filters</p>
          </div>
        )}
      </div>

      {/* Pagination */}
      {filteredTasks.length > limit && (
        <div className="mt-6 flex items-center justify-between bg-white rounded-lg shadow-sm border border-gray-200 px-6 py-4">
          <div className="text-sm text-gray-600">
            Showing <span className="font-semibold text-gray-900">{offset + 1}</span> to{' '}
            <span className="font-semibold text-gray-900">{Math.min(offset + limit, filteredTasks.length)}</span> of{' '}
            <span className="font-semibold text-gray-900">{filteredTasks.length}</span> tasks
          </div>
          
          <div className="flex items-center gap-3">
            <button
              onClick={() => setOffset(Math.max(0, offset - limit))}
              disabled={offset === 0}
              className={`px-5 py-2.5 text-sm font-semibold rounded-lg border-2 transition-all ${
                offset === 0
                  ? 'bg-gray-100 text-gray-400 border-gray-200 cursor-not-allowed'
                  : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50 hover:border-blue-500 hover:text-blue-600'
              }`}
            >
              ← Previous
            </button>
            
            <span className="text-sm font-semibold text-gray-700 px-4">
              Page {currentPage} of {totalPages}
            </span>
            
            <button
              onClick={() => setOffset(offset + limit)}
              disabled={offset + limit >= filteredTasks.length}
              className={`px-5 py-2.5 text-sm font-semibold rounded-lg border-2 transition-all ${
                offset + limit >= filteredTasks.length
                  ? 'bg-gray-100 text-gray-400 border-gray-200 cursor-not-allowed'
                  : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50 hover:border-blue-500 hover:text-blue-600'
              }`}
            >
              Next →
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
