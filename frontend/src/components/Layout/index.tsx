import { type ReactNode } from 'react'
import { Layout, Menu, Button, Popconfirm, Space } from 'antd'
import { useNavigate, useLocation } from 'react-router-dom'
import { LogoutOutlined, UserOutlined } from '@ant-design/icons'
import { useAuth } from '@/context/AuthContext'
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
  const { user, logout } = useAuth()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

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
        <Header className={styles.header}>
          {user && (
            <div className={styles.userInfo}>
              <Space className={styles.userText}>
                <UserOutlined />
                <span>欢迎您，</span>
                <span className={styles.username}>{user.username}</span>
                <span className={styles.roleTag}>{user.role}</span>
              </Space>
              
              <Popconfirm
                title="确定退出登录吗？"
                onConfirm={handleLogout}
                okText="确定"
                cancelText="取消"
                placement="bottomRight"
              >
                <Button 
                  type="text" 
                  danger 
                  icon={<LogoutOutlined />} 
                  className={styles.logoutBtn}
                >
                  退出登录
                </Button>
              </Popconfirm>
            </div>
          )}
        </Header>
        <Content className={styles.content}>{children}</Content>
      </Layout>
    </Layout>
  )
}

export default AppLayout

