import { API_BASE } from '../services/api'
import { useAuth } from '../context/AuthContext'

export function useApi() {
  const { token } = useAuth()

  const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
    const headers = {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    }

    const fullUrl = url.startsWith('http') ? url : `${API_BASE}${url}`
    const response = await fetch(fullUrl, { ...options, headers })

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }))
      throw new Error(error.error || `HTTP ${response.status}`)
    }

    return response.json()
  }

  return { fetchWithAuth }
}
