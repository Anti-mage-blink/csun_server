import { Routes, Route, Navigate } from 'react-router-dom'
import QuoteCreate from '@/modules/quote-create'

// 路由表：新增模块时在此追加一条 Route 即可，与左侧菜单 menuItems 保持对应
const AppRoutes = () => {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/quote-create" replace />} />
      <Route path="/quote-create" element={<QuoteCreate />} />
    </Routes>
  )
}

export default AppRoutes
