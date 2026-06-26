import { type ReactNode } from 'react'
import { Layout, Menu } from 'antd'
import { useNavigate, useLocation } from 'react-router-dom'
import styles from './index.module.css'

const { Sider, Content, Header } = Layout

/**
 * 左侧菜单配置
 * 新增模块时：在此追加一项，并在 routes/index.tsx 追加对应路由
 */
const menuItems = [
  { key: '/quote-create', label: '新建报价单' },
]

interface AppLayoutProps {
  children: ReactNode
}

/**
 * 整体布局：左侧纵向菜单 + 右侧内容区
 * 点击左侧菜单项，右侧切换到对应子页面（由路由驱动）
 */
const AppLayout = ({ children }: AppLayoutProps) => {
  const navigate = useNavigate()
  const location = useLocation()

  return (
    <Layout className={styles.layout}>
      <Sider>
        <div className={styles.logo}>报价管理系统 v2</div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header className={styles.header} />
        <Content className={styles.content}>{children}</Content>
      </Layout>
    </Layout>
  )
}

export default AppLayout
