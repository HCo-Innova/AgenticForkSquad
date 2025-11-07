import type { Task } from '../services/tasks'

function badgeColor(status: string) {
  switch (status) {
    case 'pending': return '#d97706' // amber
    case 'in_progress': return '#2563eb' // blue
    case 'completed': return '#16a34a' // green
    case 'failed': return '#dc2626' // red
    default: return '#6b7280' // gray
  }
}

export default function TaskCard({ task, onClick }: { task: Task; onClick?: () => void }) {
  return (
    <div onClick={onClick} style={{ border: '1px solid #eee', borderRadius: 8, padding: 12, display: 'grid', gap: 6, cursor: onClick ? 'pointer' : 'default' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <strong>#{task.id}</strong>
        <span style={{ background: badgeColor(task.status), color: 'white', borderRadius: 12, padding: '2px 8px', fontSize: 12 }}>
          {task.status.replace('_',' ')}
        </span>
      </div>
      <div>Type: {task.type}</div>
      <div style={{ color: '#6b7280', fontFamily: 'monospace', fontSize: 12 }}>Query: {task.target_query}</div>
      <div style={{ color: '#6b7280', fontSize: 12 }}>Created: {task.created_at}</div>
    </div>
  )
}
