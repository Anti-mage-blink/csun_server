import { type ReactNode, useState } from 'react'
import { Layout, Menu, Button, Popconfirm, Space } from 'antd'
import { useNavigate, useLocation } from 'react-router-dom'
import { LogoutOutlined, UserOutlined, LeftOutlined, RightOutlined } from '@ant-design/icons'
import { useAuth } from '@/AuthContext'
import { getMenuItemsByRole } from '@/roleMenuConfig'
import styles from './index.module.css'

const { Sider, Content, Header } = Layout

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
  const [collapsed, setCollapsed] = useState(false)

  // 读取当前登录用户，根据角色功能映射给出功能子页面（顺序符合映射列表顺序）
  const menuItems = getMenuItemsByRole(user?.role)

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <Layout className={styles.layout}>
      <Sider
        collapsible
        collapsed={collapsed}
        trigger={null}
        collapsedWidth={0}
        width={200}
        className="no-print"
      >
        <div className={styles.logo}>报价管理系统 v2</div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>

      {/* 收起/展开圆形按钮 */}
      <div
        className={`${styles.trigger} no-print`}
        style={{ left: collapsed ? '12px' : '200px' }}
        onClick={() => setCollapsed(!collapsed)}
        title={collapsed ? '展开侧边栏' : '收起侧边栏'}
      >
        {collapsed ? <RightOutlined /> : <LeftOutlined />}
      </div>

      <Layout>
        <Header className={`${styles.header} no-print`}>
          {user && (
            <div className={styles.userInfo}>
              <Space className={styles.userText}>
                <UserOutlined />
                <span>
                  欢迎您，<span className={styles.username}>{user.name}</span>
                </span>
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

