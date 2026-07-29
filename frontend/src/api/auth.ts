import request from './request'
import { type User } from '@/AuthContext'

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
    
    console.log('[Login API Debug] 收到后端返回原始数据:', backendData)
    
    // 2. 💡 将后端的 Employee 模型转换适配为前端期望的 LoginResponse 契约
    const mappedUser: User = {
      id: backendData.data.id,
      name: backendData.data.name || username,
      role: backendData.data.role || '普通员工', // 使用后端返回的特定角色字段
      wecomId: '' // 暂无
    }
    
    console.log('[Login API Debug] 适配转换后的前端 User 对象:', mappedUser)
    
    return {
      code: res.status, // 补齐前端需要的成功状态码
      message: backendData.message || '登录成功',
      data: {
        token: 'token-placeholder-xyz-123456', // 暂存本地 Token 占位符
        user: mappedUser
      }
    }
  } catch (error: any) {
    // 3. 💡 若后端服务在运行，但账号/密码错误（401等），直接抛出后端响应的真实错误原因
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '登录校验失败')
    }
    throw new Error(error.message || '连接登录服务器失败，请检查服务是否启动')
  }
}

/**
 * 获取所有可供密码重置选择的用户列表
 */
export const getUsersListApi = async (): Promise<User[]> => {
  try {
    const res = await request.get<User[]>('/auth/users')
    return res.data
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '获取用户列表失败')
    }
    throw new Error(error.message || '连接服务器失败，获取用户列表失败')
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
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '提交重置申请失败')
    }
    throw new Error(error.message || '连接服务器失败，提交重置申请失败')
  }
}
