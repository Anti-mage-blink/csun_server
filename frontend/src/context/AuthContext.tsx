import React, { createContext, useContext, useState, useEffect } from 'react'

export interface User {
  id: number
  username: string
  role: string
  wecomId?: string
}

interface AuthContextType {
  user: User | null
  login: (userInfo: User, token: string) => void
  logout: () => void
  loading: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

// 初始模拟用户数据，如果本地没有则初始化
const DEFAULT_USERS = [
  { id: 1, username: 'admin', password: '123456', role: '管理员', wecomId: 'admin_wecom' },
  { id: 2, username: 'sales1', password: '123456', role: '商务员', wecomId: 'sales1_wecom' },
  { id: 3, username: 'sales2', password: '123456', role: '报价员', wecomId: 'sales2_wecom' },
]

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // 初始化本地存储的模拟用户，铺垫给未来的重置密码和登录功能
    if (!localStorage.getItem('sys_users')) {
      localStorage.setItem('sys_users', JSON.stringify(DEFAULT_USERS))
    }
    
    // 初始化密码重置请求列表
    if (!localStorage.getItem('passwordResetRequests')) {
      localStorage.setItem('passwordResetRequests', JSON.stringify([]))
    }

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
