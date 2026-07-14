import request from './request'
import { type User } from '@/context/AuthContext'

export interface LoginResponse {
  code: number
  message: string
  data: {
    token: string
    user: User
  }
}

/**
 * 登录 API
 * @param username 用户名
 * @param password 密码
 * 
 * 架构设计说明（给后端对接铺地）：
 * 在后端 Go (Gin) 接口未实现时，此处采用本地 mock 进行降级处理，
 * 接口实现后，直接开启 /api/login 的网络请求，即可无缝对接。
 */
export const loginApi = async (username: string, password: string): Promise<LoginResponse> => {
  try {
    // 1. 发起真实的后端连接测试
    const res = await request.post<any>('/login', { username, password })
    const backendData = res.data // 后端真实的 JSON 返回体：{ message: string, data: Employee }
    
    // 2. 💡 将后端的 Employee 模型转换适配为前端期望的 LoginResponse 契约
    return {
      code: res.status, // 补齐前端需要的成功状态码
      message: backendData.message || '登录成功',
      data: {
        token: 'token-placeholder-xyz-123456', // 暂存本地 Token 占位符
        user: {
          id: backendData.data.id,
          username: backendData.data.name || backendData.data.employee_number || username,
          role: backendData.data.department || '普通员工', // 将后端的部门充当用户的 Role
          wecomId: '' // 暂无
        }
      }
    }
  } catch (error: any) {
    // 3. 💡 若后端服务在运行，但账号/密码错误（401等），直接抛出后端响应的真实错误原因
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '登录校验失败')
    }
    
    // 4. 当本地后端未启动或网络不通时，才走自动降级 Mock 调试（原降级逻辑保持不变）
    console.warn('后端服务未启动或连接失败，已自动降级至前端 Mock 运行环境')
    
    const usersStr = localStorage.getItem('sys_users') || '[]'
    const users = JSON.parse(usersStr)
    const found = users.find((x: any) => x.username === username && x.password === password)
    
    if (found) {
      return {
        code: 200,
        message: '登录成功',
        data: {
          token: 'mock-token-xyz-123456',
          user: {
            id: found.id,
            username: found.username,
            role: found.role,
            wecomId: found.wecomId,
          }
        }
      }
    } else {
      throw new Error('账号或密码错误')
    }
  }
}

/**
 * 获取所有可供密码重置选择的用户列表
 */
export const getUsersListApi = async (): Promise<User[]> => {
  try {
    const res = await request.get<User[]>('/auth/users')
    return res.data
  } catch (error) {
    const usersStr = localStorage.getItem('sys_users') || '[]'
    const users = JSON.parse(usersStr)
    // 排除管理员本身，或者返回全部
    return users.map((u: any) => ({
      id: u.id,
      username: u.username,
      role: u.role,
      wecomId: u.wecomId,
    }))
  }
}

/**
 * 提交忘记密码/重置申请
 * @param userId 用户 ID
 */
export const submitForgotPasswordApi = async (userId: number): Promise<{ message: string }> => {
  try {
    const res = await request.post<{ message: string }>('/auth/forgot-password', { userId })
    return res.data
  } catch (error) {
    console.warn('后端接口未开启，自动记录至 localStorage')
    
    let resetRequests = JSON.parse(localStorage.getItem('passwordResetRequests') || '[]')
    // 清除该用户之前待处理的请求
    resetRequests = resetRequests.filter((r: any) => !(r.userId === userId && r.status === 'pending'))
    
    resetRequests.push({
      id: resetRequests.length > 0 ? Math.max(...resetRequests.map((r: any) => r.id)) + 1 : 1,
      userId: userId,
      status: 'pending',
      requestedAt: new Date().toISOString(),
      requesterName: '用户自助申请'
    })
    
    localStorage.setItem('passwordResetRequests', JSON.stringify(resetRequests))
    return { message: '提交申请成功，等待管理员重置' }
  }
}
