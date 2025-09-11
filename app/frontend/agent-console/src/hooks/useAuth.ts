import { useState, useEffect } from 'react'
import { apiClient, LoginRequest, User } from '@/lib/api'

export function useAuth() {
  const [user, setUser] = useState<User | null>(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    // Check if user is already logged in
    const token = localStorage.getItem('auth_token')
    const refreshToken = localStorage.getItem('refresh_token')
    const userData = localStorage.getItem('user_data')
    // const projectId = localStorage.getItem('project_id')
    
    if (token && refreshToken && userData) {
      try {
        const parsedUser = JSON.parse(userData)
        setUser(parsedUser)
        setIsAuthenticated(true)
        apiClient.setTenantId(parsedUser.tenant_id)
        
        // Set default project if not already set
        // if (projectId) {
        //   apiClient.setProjectId(projectId)
        // } else {
        //   apiClient.setProjectId('550e8400-e29b-41d4-a716-446655440001')
        // }
      } catch (_err) {
        // Clear corrupted data
        localStorage.removeItem('auth_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('user_data')
        localStorage.removeItem('tenant_id')
        localStorage.removeItem('project_id')
      }
    } else if (refreshToken && userData) {
      // Try to refresh the token if we have a refresh token but no access token
      try {
        const parsedUser = JSON.parse(userData)
        apiClient.setTenantId(parsedUser.tenant_id)
        // if (projectId) {
        //   apiClient.setProjectId(projectId)
        // } else {
        //   apiClient.setProjectId('550e8400-e29b-41d4-a716-446655440001')
        // }
        apiClient.refreshToken().then(() => {
          setUser(parsedUser)
          setIsAuthenticated(true)
        }).catch(() => {
          // Refresh failed, clear everything
          localStorage.removeItem('auth_token')
          localStorage.removeItem('refresh_token')
          localStorage.removeItem('user_data')
          localStorage.removeItem('tenant_id')
          localStorage.removeItem('project_id')
        })
      } catch (_err) {
        // Clear corrupted data
        localStorage.removeItem('auth_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('user_data')
        localStorage.removeItem('tenant_id')
        localStorage.removeItem('project_id')
      }
    }
    
    setIsLoading(false)
  }, [])

  const login = async (credentials: LoginRequest) => {
    setIsLoading(true)
    setError(null)
    
    try {
      const response = await apiClient.login(credentials)
      setUser(response.user)
      setIsAuthenticated(true)
      localStorage.setItem('user_data', JSON.stringify(response.user))
      apiClient.setTenantId(response.user.tenant_id)
      
      setIsLoading(false)
      
      // Navigate to inbox after successful login
      window.location.assign("/tickets")
    } catch (error: any) {
      setError(error.response?.data?.message || 'Login failed')
      setIsLoading(false)
      throw error
    }
  }

  const logout = () => {
    apiClient.logout()
    setUser(null)
    setIsAuthenticated(false)
    window.location.assign("/login")
  }

  const clearError = () => {
    setError(null)
  }

  return {
    user,
    isAuthenticated,
    isLoading,
    error,
    login,
    logout,
    clearError,
  }
}
