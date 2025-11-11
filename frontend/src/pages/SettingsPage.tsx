import { useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { useTheme } from '../context/ThemeContext'
import toast from 'react-hot-toast'
import { apiGet } from '../services/api'

export default function SettingsPage() {
  const { user } = useAuth()
  const { theme, toggleTheme } = useTheme()
  const [notifications, setNotifications] = useState(
    localStorage.getItem('notifications') !== 'false'
  )
  const [autoRefresh, setAutoRefresh] = useState(
    localStorage.getItem('autoRefresh') !== 'false'
  )
  const [refreshInterval, setRefreshInterval] = useState(
    parseInt(localStorage.getItem('refreshInterval') || '30')
  )
  const [showDeleteModal, setShowDeleteModal] = useState(false)

  const handleThemeToggle = () => {
    toggleTheme()
    toast.success(`Theme changed to ${theme === 'light' ? 'dark' : 'light'}`)
  }

  const handleNotificationsToggle = () => {
    const newValue = !notifications
    setNotifications(newValue)
    localStorage.setItem('notifications', String(newValue))
    toast.success(newValue ? 'Notifications enabled' : 'Notifications disabled')
  }

  const handleAutoRefreshToggle = () => {
    const newValue = !autoRefresh
    setAutoRefresh(newValue)
    localStorage.setItem('autoRefresh', String(newValue))
    toast.success(newValue ? 'Auto-refresh enabled' : 'Auto-refresh disabled')
  }

  const handleRefreshIntervalChange = (value: number) => {
    setRefreshInterval(value)
    localStorage.setItem('refreshInterval', String(value))
    toast.success(`Refresh interval set to ${value} seconds`)
  }

  const handleExportData = async () => {
    try {
      const toastId = toast.loading('Exporting data...')
      
      // Fetch all data in parallel
      const [tasks, agents, analytics] = await Promise.all([
        apiGet<any>('/api/v1/tasks'),
        apiGet<any>('/api/v1/agents'),
        apiGet<any>('/api/v1/metrics/overview')
      ])

      // Create export object
      const exportData = {
        exported_at: new Date().toISOString(),
        user: user?.email,
        tasks: tasks?.data || [],
        agents: agents?.data || [],
        analytics: analytics?.data || {}
      }

      // Download as JSON file
      const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `afs-export-${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)

      toast.success('Data exported successfully', { id: toastId })
    } catch (error) {
      toast.error('Failed to export data')
    }
  }

  const handleClearCache = () => {
    // Clear some localStorage items but preserve auth
    const token = localStorage.getItem('token')
    const userStr = localStorage.getItem('user')
    localStorage.clear()
    if (token) localStorage.setItem('token', token)
    if (userStr) localStorage.setItem('user', userStr)
    toast.success('Cache cleared successfully')
  }

  return (
    <div className="max-w-4xl mx-auto py-6 space-y-6">
      <h1 className="text-3xl font-bold text-gray-900">Settings</h1>

      {/* User Profile Section */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">User Profile</h2>
        </div>
        <div className="px-6 py-4 space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Full Name</label>
              <div className="mt-1 text-sm text-gray-900">{user?.full_name || 'N/A'}</div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Email</label>
              <div className="mt-1 text-sm text-gray-900">{user?.email}</div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Role</label>
              <div className="mt-1">
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                  user?.role === 'admin' ? 'bg-purple-100 text-purple-800' : 'bg-blue-100 text-blue-800'
                }`}>
                  {user?.role}
                </span>
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Status</label>
              <div className="mt-1">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                  Active
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Appearance Settings */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Appearance</h2>
        </div>
        <div className="px-6 py-4 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Theme</label>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">Current: {theme}</span>
              <button
                onClick={handleThemeToggle}
                className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                  theme === 'dark' ? 'bg-blue-600' : 'bg-gray-200'
                }`}
              >
                <span
                  className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    theme === 'dark' ? 'translate-x-6' : 'translate-x-1'
                  }`}
                />
              </button>
              <span className="text-sm text-gray-600">{theme === 'light' ? '☀️ Light' : '🌙 Dark'}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Notifications Settings */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Notifications</h2>
        </div>
        <div className="px-6 py-4 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-900">Enable Notifications</div>
              <div className="text-xs text-gray-500">Receive toast notifications for events</div>
            </div>
            <button
              onClick={handleNotificationsToggle}
              className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                notifications ? 'bg-blue-600' : 'bg-gray-200'
              }`}
            >
              <span
                className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                  notifications ? 'translate-x-6' : 'translate-x-1'
                }`}
              />
            </button>
          </div>
        </div>
      </div>

      {/* Dashboard Settings */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Dashboard</h2>
        </div>
        <div className="px-6 py-4 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-900">Auto-refresh</div>
              <div className="text-xs text-gray-500">Automatically refresh dashboard data</div>
            </div>
            <button
              onClick={handleAutoRefreshToggle}
              className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                autoRefresh ? 'bg-blue-600' : 'bg-gray-200'
              }`}
            >
              <span
                className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                  autoRefresh ? 'translate-x-6' : 'translate-x-1'
                }`}
              />
            </button>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Refresh Interval: {refreshInterval} seconds
            </label>
            <input
              type="range"
              min="10"
              max="120"
              step="10"
              value={refreshInterval}
              onChange={(e) => handleRefreshIntervalChange(parseInt(e.target.value))}
              className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
            />
            <div className="flex justify-between text-xs text-gray-500 mt-1">
              <span>10s</span>
              <span>30s</span>
              <span>60s</span>
              <span>120s</span>
            </div>
          </div>
        </div>
      </div>

      {/* Data Management */}
      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Data Management</h2>
        </div>
        <div className="px-6 py-4 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-900">Export Data</div>
              <div className="text-xs text-gray-500">Download your task and analytics data</div>
            </div>
            <button
              onClick={handleExportData}
              className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 transition-colors"
            >
              Export
            </button>
          </div>

          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-900">Clear Cache</div>
              <div className="text-xs text-gray-500">Clear local storage cache</div>
            </div>
            <button
              onClick={handleClearCache}
              className="px-4 py-2 bg-gray-600 text-white text-sm font-medium rounded-md hover:bg-gray-700 transition-colors"
            >
              Clear
            </button>
          </div>
        </div>
      </div>

      {/* Danger Zone */}
      <div className="bg-white rounded-lg shadow border-2 border-red-200">
        <div className="px-6 py-4 border-b border-red-200 bg-red-50">
          <h2 className="text-xl font-semibold text-red-900">Danger Zone</h2>
        </div>
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-900">Delete Account</div>
              <div className="text-xs text-gray-500">Permanently delete your account and all data</div>
            </div>
            <button
              onClick={() => setShowDeleteModal(true)}
              className="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-md hover:bg-red-700 transition-colors"
            >
              Delete
            </button>
          </div>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      {showDeleteModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900">Delete Account</h3>
            </div>
            <div className="px-6 py-4">
              <p className="text-sm text-gray-700 mb-4">
                This action <strong>cannot be undone</strong>. This will permanently delete your account,
                all tasks, agent executions, and analytics data.
              </p>
              <p className="text-sm text-red-600 font-medium">
                Are you absolutely sure?
              </p>
            </div>
            <div className="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
              <button
                onClick={() => setShowDeleteModal(false)}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => {
                  toast.error('Account deletion requires backend endpoint implementation')
                  setShowDeleteModal(false)
                }}
                className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700 transition-colors"
              >
                Yes, Delete My Account
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
