import { Routes, Route } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'
import { AuthProvider } from './context/AuthContext'
import { ThemeProvider } from './context/ThemeContext'
import ProtectedRoute from './components/ProtectedRoute'
import Nav from './components/Nav'
import HomePage from './pages/HomePage'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import TaskDetailPage from './pages/TaskDetailPage'
import AgentsPage from './pages/AgentsPage'
import TaskSubmissionPage from './pages/TaskSubmissionPage'
import TaskListPage from './pages/TaskListPage'
import AboutPage from './pages/AboutPage'
import AnalyticsPage from './pages/AnalyticsPage'
import SettingsPage from './pages/SettingsPage'

export default function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <Toaster position="top-right" />
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          
          <Route path="/*" element={
            <ProtectedRoute>
              <div>
                <Nav />
                <div style={{ padding: '1rem' }}>
                  <Routes>
                    <Route path="/" element={<HomePage />} />
                    <Route path="/tasks" element={<TaskListPage />} />
                    <Route path="/tasks/new" element={<TaskSubmissionPage />} />
                    <Route path="/tasks/:id" element={<TaskDetailPage />} />
                    <Route path="/agents" element={<AgentsPage />} />
                    <Route path="/about" element={<AboutPage />} />
                    <Route path="/analytics" element={<AnalyticsPage />} />
                    <Route path="/settings" element={<SettingsPage />} />
                  </Routes>
                </div>
              </div>
            </ProtectedRoute>
          } />
        </Routes>
      </AuthProvider>
    </ThemeProvider>
  )
}