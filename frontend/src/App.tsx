import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import AppLayout from '@/components/Layout'
import Login from '@/pages/login'
import CreateQuote from '@/pages/create-quote'
import Filing from '@/pages/filing'
import MyApplications from '@/pages/my-applications'
import MyApprovals from '@/pages/my-approvals'
import { AuthProvider, useAuth } from '@/context/AuthContext'
import { Spin } from 'antd'

function AppRoutes() {
  const { user } = useAuth()

  // 根据角色自动确定默认跳转的页面
  const getDefaultRedirect = () => {
    if (!user) return '/login'
    
    // 管理员角色默认拥有所有功能，重定向至新建报价单或首个可用功能
    if (user.role === '管理员' || user.role === '上帝') {
      return '/create-quote'
    }
    
    if (user.role === '市场部') {
      return '/create-quote'
    }
    if (user.role === '财务部') {
      return '/filing'
    }
    if (
      user.role === '领导小组副组长' || 
      user.role === '领导小组组长' || 
      user.role === '工作小组组长-光伏热场' || 
      user.role === '工作小组组长-摩擦'
    ) {
      return '/my-approvals'
    }
    return '/create-quote' // 其它未知角色默认跳转
  }

  return (
    <Routes>
      <Route path="/" element={<Navigate to={getDefaultRedirect()} replace />} />
      <Route path="/create-quote" element={<CreateQuote />} />
      <Route path="/filing" element={<Filing />} />
      <Route path="/my-applications" element={<MyApplications />} />
      <Route path="/my-approvals" element={<MyApprovals />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

function AppContent() {
  const { user, loading } = useAuth()

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', background: '#f5f5f5' }}>
        <Spin size="large" tip="系统初始化中..." />
      </div>
    )
  }

  return (
    <BrowserRouter>
      {user ? (
        <AppLayout>
          <AppRoutes />
        </AppLayout>
      ) : (
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      )}
    </BrowserRouter>
  )
}

function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  )
}

export default App

