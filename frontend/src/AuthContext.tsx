import React, { createContext, useContext, useState, useEffect } from 'react'

export interface User {
  id: number
  name: string
  role: string
}

interface AuthContextType {
  user: User | null
  login: (userInfo: User, token: string) => void
  logout: () => void
  loading: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const storedUser = localStorage.getItem('currentUser')
    const token = localStorage.getItem('token')

    if (storedUser && token) {
      try {
        setUser(JSON.parse(storedUser))
      } catch (e) {
        // 解析失败则清除
        localStorage.removeItem('currentUser')
        localStorage.removeItem('token')
      }
    }
    setLoading(false)
  }, [])

  const login = (userInfo: User, token: string) => {
    localStorage.setItem('currentUser', JSON.stringify(userInfo))
    localStorage.setItem('token', token)
    setUser(userInfo)
  }

  const logout = () => {
    localStorage.removeItem('currentUser')
    localStorage.removeItem('token')
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, login, logout, loading }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
