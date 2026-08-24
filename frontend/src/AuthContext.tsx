import React, { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { myApproveQueryApi } from '@/api/quote'

export interface User {
  id: number
  name: string
  role: string
}

// 判断是否为审批角色
export const isApproverRole = (role?: string): boolean => {
  if (!role) return false
  const approverRoles = [
    '工作小组组长-光伏热场',
    '工作小组组长-摩擦',
    '领导小组副组长',
    '领导小组组长',
    '组长',
    '副组长',
    '系统管理员'
  ]
  return approverRoles.includes(role)
}

interface AuthContextType {
  user: User | null
  login: (userInfo: User, token: string) => void
  logout: () => void
  loading: boolean
  pendingCount: number | null
  fetchPendingCount: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [pendingCount, setPendingCount] = useState<number | null>(null)

  // 查待审批数据并更新待审批数量
  const fetchPendingCount = useCallback(async () => {
    const storedUser = localStorage.getItem('currentUser')
    const currentUser = user || (storedUser ? JSON.parse(storedUser) : null)
    if (!currentUser || !isApproverRole(currentUser.role)) {
      setPendingCount(null)
      return
    }
    try {
      const res = await myApproveQueryApi(currentUser.id)
      const processes = res.data?.quote_processes || []
      const nodes = res.data?.quote_process_nodes || []

      const count = processes.filter((p) => {
        const userNodes = nodes.filter(
          (n) => n.process_id === p.id && n.approver_id === currentUser.id
        )
        return userNodes.some((n) => n.status === '待审批')
      }).length

      setPendingCount(count)
    } catch (err) {
      console.error('自动拉取待审批数量失败:', err)
    }
  }, [user])

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

  useEffect(() => {
    if (user && isApproverRole(user.role)) {
      fetchPendingCount()
    } else {
      setPendingCount(null)
    }
  }, [user, fetchPendingCount])

  const login = (userInfo: User, token: string) => {
    localStorage.setItem('currentUser', JSON.stringify(userInfo))
    localStorage.setItem('token', token)
    setUser(userInfo)
  }

  const logout = () => {
    localStorage.removeItem('currentUser')
    localStorage.removeItem('token')
    setUser(null)
    setPendingCount(null)
  }

  return (
    <AuthContext.Provider value={{ user, login, logout, loading, pendingCount, fetchPendingCount }}>
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
