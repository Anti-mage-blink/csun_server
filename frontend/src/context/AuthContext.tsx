import React, { createContext, useContext, useState, useEffect } from 'react'
import { type User } from '@/types/gorm-models'

export type { User }

interface AuthContextType {
  user: User | null
  login: (userInfo: User, token: string) => void
  logout: () => void
  loading: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

// 模拟员工表数据（对应 backend/dao/model/general/employee.gen.go 中的 Employee 结构）
const DEFAULT_EMPLOYEES = [
  { id: 1, employee_number: 'EMP001', name: '张三（管理员）', department: '管理部' },
  { id: 2, employee_number: 'EMP002', name: '李四（商务员）', department: '商务部' },
  { id: 3, employee_number: 'EMP003', name: '王五（报价员）', department: '报价部' },
]

// 模拟账户表数据（对应 backend/dao/model/general/account.gen.go 中的 Account 结构，通过 employee_id 关联 Employee）
const DEFAULT_ACCOUNTS = [
  { id: 1, employee_id: 1, username: 'admin', password: '123456' },
  { id: 2, employee_id: 2, username: 'sales1', password: '123456' },
  { id: 3, employee_id: 3, username: 'sales2', password: '123456' },
]

// 模拟报价员工表数据（对应 backend/dao/model/quote_manage/quote_employee.gen.go 结构）
const DEFAULT_QUOTE_EMPLOYEES = [
  { id: 1, employee_id: 1, role: '管理员' },
  { id: 2, employee_id: 2, role: '商务员' },
  { id: 3, employee_id: 3, role: '报价员' },
]

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // 初始化本地存储的模拟数据库表，铺垫给未来的多表关联和登录逻辑
    if (!localStorage.getItem('sys_employees')) {
      localStorage.setItem('sys_employees', JSON.stringify(DEFAULT_EMPLOYEES))
    }
    if (!localStorage.getItem('sys_accounts')) {
      localStorage.setItem('sys_accounts', JSON.stringify(DEFAULT_ACCOUNTS))
    }
    if (!localStorage.getItem('sys_quote_employees')) {
      localStorage.setItem('sys_quote_employees', JSON.stringify(DEFAULT_QUOTE_EMPLOYEES))
    }

    // 同时初始化并保持原有的 sys_users 兼容层，保证原有密码重置申请逻辑能平滑运行
    if (!localStorage.getItem('sys_users')) {
      const legacyUsers = DEFAULT_ACCOUNTS.map(acc => {
        const qEmp = DEFAULT_QUOTE_EMPLOYEES.find(q => q.employee_id === acc.employee_id)
        return {
          id: acc.id,
          username: acc.username,
          password: acc.password,
          role: qEmp ? qEmp.role : '普通员工',
          wecomId: undefined
        }
      })
      localStorage.setItem('sys_users', JSON.stringify(legacyUsers))
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
