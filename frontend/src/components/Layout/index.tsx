import { type ReactNode, useState } from 'react'
import { Layout, Menu, Button, Popconfirm, Space } from 'antd'
import { useNavigate, useLocation } from 'react-router-dom'
import { LogoutOutlined, UserOutlined, LeftOutlined, RightOutlined } from '@ant-design/icons'
import { useAuth } from '@/context/AuthContext'
import styles from './index.module.css'

const { Sider, Content, Header } = Layout

// 定义所有的菜单项及其对应的角色权限
const ALL_MENU_ITEMS = [
  { key: '/create-quote', label: '新建报价单', roles: ['市场部', '管理员','工作小组组长-光伏热场', '工作小组组长-摩擦', '上帝'] },
  { key: '/filing', label: '备案查看', roles: ['财务部', '管理员', '上帝'] },
  { key: '/my-applications', label: '我的申请', roles: ['市场部', '上帝'] },
  { key: '/my-approvals', label: '我的审批', roles: ['领导小组组长', '领导小组副组长', '工作小组组长-光伏热场', '工作小组组长-摩擦', '上帝'] },
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
  const [collapsed, setCollapsed] = useState(false)

  // 根据当前用户的角色过滤菜单项
  const menuItems = ALL_MENU_ITEMS.filter(item => {
    if (!user) return false
    return item.roles.includes(user.role)
  })

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

