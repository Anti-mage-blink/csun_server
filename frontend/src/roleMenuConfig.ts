import type { MenuProps } from 'antd'

// 菜单项配置类型，直接复用 Ant Design Menu 的 ItemType
export type MenuItemConfig = NonNullable<MenuProps['items']>[number]

// 所有功能子页面的元数据配置
export const MENU_ITEM_MAP: Record<string, MenuItemConfig> = {
  '/create-quote': { key: '/create-quote', label: '新建报价单' },
  '/filing': { key: '/filing', label: '备案查看' },
  '/my-applications': { key: '/my-applications', label: '我的申请' },
  '/my-approvals': { key: '/my-approvals', label: '我的审批' },
  '/product-manage': { key: '/product-manage', label: '产品管理' },
  '/test-page': { key: '/test-page', label: '测试页面' },
}

/**
 * 角色 - 功能映射：角色为键，功能 key 数组为值
 * 列表中的顺序即为菜单栏中的展示顺序以及默认跳转的优先级顺序
 */
export const ROLE_FUNCTIONS_MAP: Record<string, string[]> = {
  '市场部': ['/create-quote', '/my-applications'],
  '财务部': ['/filing', '/product-manage'],
  '领导小组组长': ['/my-approvals', '/filing'],
  '领导小组副组长': ['/my-approvals', '/filing'],
  '工作小组组长-光伏热场': ['/my-approvals'],
  '工作小组组长-摩擦': ['/my-approvals'],
  '上帝': ['/create-quote', '/filing', '/my-applications', '/my-approvals', '/product-manage', '/test-page'],
}

/**
 * 读取当前登录用户的角色，根据角色功能映射，给出功能子页面
 * 功能子页面的顺序符合 ROLE_FUNCTIONS_MAP 中该角色的功能子页面列表中的顺序
 */
export const getMenuItemsByRole = (role?: string): MenuItemConfig[] => {
  if (!role) return []
  const functionKeys = ROLE_FUNCTIONS_MAP[role] || []
  return functionKeys
    .map((key) => MENU_ITEM_MAP[key])
    .filter((item): item is MenuItemConfig => Boolean(item))
}
