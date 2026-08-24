import { type ReactNode, useState } from 'react'
import { Layout, Menu, Button, Popconfirm, Space, Drawer } from 'antd'
import { useNavigate, useLocation } from 'react-router-dom'
import { LogoutOutlined, UserOutlined, LeftOutlined, RightOutlined, MenuOutlined } from '@ant-design/icons'
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
  const { user, logout, pendingCount } = useAuth()
  const [collapsed, setCollapsed] = useState(false)
  const [mobileDrawerOpen, setMobileDrawerOpen] = useState(false)

  // 读取当前登录用户，根据角色功能映射给出功能子页面（顺序符合映射列表顺序）
  const baseMenuItems = getMenuItemsByRole(user?.role)

  const menuItems = baseMenuItems.map((item) => {
    if (item && item.key === '/my-approvals') {
      return {
        ...item,
        label: (
          <span style={{ display: 'inline-flex', alignItems: 'center' }}>
            <span>我的审批</span>
            {typeof pendingCount === 'number' && pendingCount > 0 && (
              <span className={styles.menuBadge}>
                {pendingCount > 99 ? '99+' : pendingCount}
              </span>
            )}
          </span>
        ),
      }
    }
    return item
  })

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <Layout className={styles.layout}>
      {/* PC 端侧边栏 */}
      <Sider
        collapsible
        collapsed={collapsed}
        trigger={null}
        collapsedWidth={0}
        width={200}
        className={`no-print ${styles.sider} ${styles.desktopSider}`}
      >
        <div className={styles.logo}>报价管理系统</div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>

      {/* 移动端 Drawer 抽屉 */}
      <Drawer
        placement="left"
        onClose={() => setMobileDrawerOpen(false)}
        open={mobileDrawerOpen}
        width={220}
        styles={{ body: { padding: 0, backgroundColor: '#2e5b88' } }}
        headerStyle={{ display: 'none' }}
        className={styles.mobileDrawer}
      >
        <div className={styles.logo}>报价管理系统</div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => {
            navigate(key)
            setMobileDrawerOpen(false)
          }}
        />
      </Drawer>

      {/* PC 端收起/展开圆形按钮 */}
      <div
        className={`${styles.trigger} ${styles.desktopTrigger} no-print`}
        style={{ left: collapsed ? '12px' : '200px' }}
        onClick={() => setCollapsed(!collapsed)}
        title={collapsed ? '展开侧边栏' : '收起侧边栏'}
      >
        {collapsed ? <RightOutlined /> : <LeftOutlined />}
      </div>

      <Layout>
        <Header className={`${styles.header} no-print`}>
          {/* 移动端汉堡包菜单按钮 */}
          <Button
            type="text"
            icon={<MenuOutlined style={{ fontSize: 18, color: '#1677ff' }} />}
            className={styles.mobileMenuBtn}
            onClick={() => setMobileDrawerOpen(true)}
          />

          {user && (
            <div className={styles.userInfo}>
              <Space className={styles.userText}>
                <UserOutlined />
                <span>
                  <span className={styles.welcomeText}>欢迎您，</span>
                  <span className={styles.username}>{user.name}</span>
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
                  <span className={styles.logoutText}>退出登录</span>
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

