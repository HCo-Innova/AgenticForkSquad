import { Routes, Route } from 'react-router-dom'
import Nav from './components/Nav'
import HomePage from './pages/HomePage'
import TaskDetailPage from './pages/TaskDetailPage'
import AgentsPage from './pages/AgentsPage'
import TaskSubmissionPage from './pages/TaskSubmissionPage'
import TaskListPage from './pages/TaskListPage'

export default function App() {
  return (
    <div>
      <Nav />
      <div style={{ padding: '1rem' }}>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/tasks" element={<TaskListPage />} />
          <Route path="/tasks/new" element={<TaskSubmissionPage />} />
          <Route path="/tasks/:id" element={<TaskDetailPage />} />
          <Route path="/agents" element={<AgentsPage />} />
        </Routes>
      </div>
    </div>
  )
}