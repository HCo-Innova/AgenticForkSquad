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
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (!form.type || !form.target_query.trim()) {
      setError('Tipo y Target Query son obligatorios')
      return
    }
    try {
      setLoading(true)
      const task = await createTask(form)
      navigate(`/tasks/${task.id}`)
    } catch (err: any) {
      setError(err?.message ?? 'Error creando tarea')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <h2>Crear Tarea</h2>
      <form onSubmit={onSubmit} style={{ display: 'grid', gap: 12, maxWidth: 640 }}>
        <label>
          <div>Tipo</div>
          <select value={form.type} onChange={(e) => onChange('type', e.target.value)}>
            <option value="query_optimization">Query Optimization</option>
          </select>
        </label>
        <label>
          <div>Descripción (opcional)</div>
          <input value={form.description ?? ''} onChange={(e) => onChange('description', e.target.value)} placeholder="Descripción" />
        </label>
        <label>
          <div>Target Query</div>
          <textarea value={form.target_query} onChange={(e) => onChange('target_query', e.target.value)} placeholder="SELECT ..." rows={6} />
        </label>
        {error && <div style={{ color: 'crimson' }}>{error}</div>}
        <button type="submit" disabled={loading}>{loading ? 'Creando...' : 'Crear'}</button>
      </form>
    </div>
  )
}
