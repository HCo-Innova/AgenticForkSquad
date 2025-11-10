import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { createTask, CreateTaskInput } from '../services/tasks'

export default function TaskSubmissionPage() {
  const navigate = useNavigate()
  const [form, setForm] = useState<CreateTaskInput>({ type: 'query_optimization', target_query: '', description: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function onChange<K extends keyof CreateTaskInput>(key: K, val: CreateTaskInput[K]) {
    setForm((f) => ({ ...f, [key]: val }))
    // Clear error when user types
    if (error) setError(null)
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    
    // Validation
    if (!form.type || !form.target_query.trim()) {
      setError('Task type and SQL query are required')
      return
    }
    
    if (form.target_query.trim().length < 10) {
      setError('SQL query is too short (minimum 10 characters)')
      return
    }
    
    try {
      setLoading(true)
      const task = await createTask(form)
      navigate(`/tasks/${task.id}`)
    } catch (err: any) {
      setError(err?.message ?? 'Failed to create task')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="mb-6">
        <h2 className="text-3xl font-bold text-gray-900">Create New Optimization Task</h2>
        <p className="mt-2 text-sm text-gray-600">
          The system will assign 3 specialized agents that will work in parallel to optimize your query.
        </p>
      </div>

      <form onSubmit={onSubmit} className="space-y-6 bg-white shadow-md rounded-lg p-6">
        {/* Task Type */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Task Type
          </label>
          <select 
            value={form.type} 
            onChange={(e) => onChange('type', e.target.value)}
            className="w-full px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
          >
            <option value="query_optimization">Query Optimization</option>
            <option value="index_tuning">Index Tuning</option>
            <option value="schema_improvement">Schema Improvement</option>
            <option value="partitioning">Partitioning</option>
          </select>
        </div>

        {/* Description */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Description (optional)
          </label>
          <input 
            type="text"
            value={form.description ?? ''} 
            onChange={(e) => onChange('description', e.target.value)} 
            placeholder="E.g., Query takes 5+ seconds with 100K orders"
            maxLength={200}
            className="w-full px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
          />
          {form.description && (
            <p className="mt-1 text-xs text-gray-500">
              {form.description.length}/200 characters
            </p>
          )}
        </div>

        {/* SQL Query */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            SQL Query to Optimize <span className="text-red-500">*</span>
          </label>
          <textarea 
            value={form.target_query} 
            onChange={(e) => onChange('target_query', e.target.value)} 
            placeholder="SELECT o.*, p.amount FROM orders o JOIN payments p ON o.id = p.order_id WHERE o.user_id = 12345 AND o.created_at > '2024-01-01' ORDER BY o.created_at DESC"
            rows={8}
            required
            className="w-full px-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm transition-all"
          />
          <div className="mt-1 flex items-center justify-between text-xs text-gray-500">
            <span>Paste the SQL query that needs optimization. The 3 agents will analyze it in parallel.</span>
            {form.target_query && (
              <span className="font-medium">
                {form.target_query.length} chars
              </span>
            )}
          </div>
        </div>

        {/* Error */}
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-md flex items-start">
            <svg className="w-5 h-5 mr-2 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
            <span>{error}</span>
          </div>
        )}

        {/* Submit Button */}
        <div className="flex gap-3 pt-2">
          <button 
            type="submit" 
            disabled={loading || !form.target_query.trim()}
            className="flex-1 bg-blue-600 text-white px-6 py-3 rounded-md font-medium hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? (
              <span className="flex items-center justify-center">
                <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Creating task...
              </span>
            ) : (
              '🚀 Create Task and Execute Agents'
            )}
          </button>
          <button 
            type="button"
            onClick={() => navigate('/tasks')}
            disabled={loading}
            className="px-6 py-3 border border-gray-300 rounded-md font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  )
}
