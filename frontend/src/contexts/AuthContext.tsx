import React, { createContext, useContext, useState, useEffect, useRef } from 'react'

interface AuthContextType {
  token: string | null
  username: string | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  changePassword: (oldPassword: string, newPassword: string) => Promise<void>
  refreshToken: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [token, setToken] = useState<string | null>(null)
  const [username, setUsername] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const refreshTokenRef = useRef<string | null>(null)
  const refreshTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    const savedToken = localStorage.getItem('auth_token')
    const savedUsername = localStorage.getItem('auth_username')
    const savedRefreshToken = localStorage.getItem('auth_refresh_token')
    
    if (savedToken && savedUsername && savedRefreshToken) {
      setToken(savedToken)
      setUsername(savedUsername)
      refreshTokenRef.current = savedRefreshToken
      scheduleTokenRefresh()
    }
    setIsLoading(false)
  }, [])

  const scheduleTokenRefresh = () => {
    if (refreshTimeoutRef.current) {
      clearTimeout(refreshTimeoutRef.current)
    }

    refreshTimeoutRef.current = setTimeout(() => {
      refreshTokenAsync()
    }, 2 * 60 * 60 * 1000 - 5 * 60 * 1000)
  }

  const refreshTokenAsync = async () => {
    try {
      await refreshToken()
    } catch (error) {
      console.error('Token refresh failed:', error)
      logout()
    }
  }

  const login = async (username: string, password: string) => {
    const response = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    })

    if (!response.ok) {
      throw new Error('Login failed')
    }

    const data = await response.json()
    setToken(data.token)
    setUsername(data.username)
    refreshTokenRef.current = data.refresh_token
    
    localStorage.setItem('auth_token', data.token)
    localStorage.setItem('auth_username', data.username)
    localStorage.setItem('auth_refresh_token', data.refresh_token)
    
    scheduleTokenRefresh()
  }

  const logout = () => {
    setToken(null)
    setUsername(null)
    refreshTokenRef.current = null
    
    if (refreshTimeoutRef.current) {
      clearTimeout(refreshTimeoutRef.current)
    }
    
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_username')
    localStorage.removeItem('auth_refresh_token')
  }

  const refreshToken = async () => {
    if (!refreshTokenRef.current) {
      throw new Error('No refresh token available')
    }

    const response = await fetch('/api/refresh-token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshTokenRef.current }),
    })

    if (!response.ok) {
      throw new Error('Token refresh failed')
    }

    const data = await response.json()
    setToken(data.token)
    refreshTokenRef.current = data.refresh_token
    
    localStorage.setItem('auth_token', data.token)
    localStorage.setItem('auth_refresh_token', data.refresh_token)
    
    scheduleTokenRefresh()
  }

  const changePassword = async (oldPassword: string, newPassword: string) => {
    const response = await fetch('/api/change-password', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
    })

    if (!response.ok) {
      throw new Error('Failed to change password')
    }
  }

  return (
    <AuthContext.Provider value={{ token, username, isAuthenticated: !!token, isLoading, login, logout, changePassword, refreshToken }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}
