import { Link } from 'react-router-dom'

export default function Nav() {
  return (
    <nav style={{ display: 'flex', gap: 12, padding: '0.75rem', borderBottom: '1px solid #eee' }}>
      <Link to="/">Home</Link>
      <Link to="/tasks">Tasks</Link>
      <Link to="/tasks/new">New Task</Link>
      <Link to="/agents">Agents</Link>
    </nav>
  )
}
