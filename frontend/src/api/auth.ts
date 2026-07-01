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
    // 铺底网络接口：尝试发起后端调用（在容器网络中为 /api/auth/login 或 /api/login）
    // 目前捕获错误，用于后端未启动时本地降级调试
    const res = await request.post<LoginResponse>('/auth/login', { username, password })
    return res.data
  } catch (error) {
    console.warn('后端服务未启动或连接失败，已自动降级至前端 Mock 运行环境')
    
    // 前端 Mock 降级逻辑
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
