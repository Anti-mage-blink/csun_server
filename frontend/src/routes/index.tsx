import { Routes, Route, Navigate } from 'react-router-dom'
import CreateQuote from '@/pages/create-quote'

// 路由表：新增模块时在此追加一条 Route 即可，与左侧菜单 menuItems 保持对应
const AppRoutes = () => {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/create-quote" replace />} />
      <Route path="/create-quote" element={<CreateQuote />} />
    </Routes>
  )
}

export default AppRoutes
