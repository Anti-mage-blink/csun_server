import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import AppLayout from '@/components/Layout'
import AppRoutes from '@/routes'
import Login from '@/modules/login'
import { AuthProvider, useAuth } from '@/context/AuthContext'
import { Spin } from 'antd'

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

