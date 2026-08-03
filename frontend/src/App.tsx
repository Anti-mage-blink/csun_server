import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import AppLayout from '@/components/Layout'
import Login from '@/pages/login'
import CreateQuote from '@/pages/create-quote'
import Filing from '@/pages/filing'
import MyApplications from '@/pages/my-applications'
import MyApprovals from '@/pages/my-approvals'
import ProductManage from '@/pages/product-manage'
import TestPage from '@/pages/test-page'
import { AuthProvider, useAuth } from '@/AuthContext'
import { Spin } from 'antd'
import { ROLE_FUNCTIONS_MAP } from '@/roleMenuConfig'

function AppRoutes() {
  const { user } = useAuth()

  // 根据角色自动确定默认跳转的页面（读取当前登录用户的角色功能映射，给出列表中第一个功能子页面）
  const getDefaultRedirect = () => {
    if (!user) return '/login'
    
    const userFunctions = ROLE_FUNCTIONS_MAP[user.role]
    if (userFunctions && userFunctions.length > 0) {
      return userFunctions[0]
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
      <Route path="/product-manage" element={<ProductManage />} />
      <Route path="/test-page" element={<TestPage />} />
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

