export type TaskFiltersValue = {
  status?: string
  type?: string
}

export default function TaskFilters({ value, onChange }: { value: TaskFiltersValue; onChange: (v: TaskFiltersValue) => void }) {
  return (
    <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 12 }}>
      <label style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
        <span>Status</span>
        <select value={value.status ?? ''} onChange={(e) => onChange({ ...value, status: e.target.value || undefined })}>
          <option value="">All</option>
          <option value="pending">Pending</option>
          <option value="in_progress">In Progress</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
        </select>
      </label>
      <label style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
        <span>Type</span>
        <select value={value.type ?? ''} onChange={(e) => onChange({ ...value, type: e.target.value || undefined })}>
          <option value="">All</option>
          <option value="query_optimization">Query Optimization</option>
        </select>
      </label>
    </div>
  )
}
