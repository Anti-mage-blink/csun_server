import axios, {
  type AxiosInstance,
  type InternalAxiosRequestConfig,
  type AxiosResponse,
} from 'axios'

/**
 * axios 实例（基础设施层）
 *
 * 说明：此处仅完成实例创建与基础配置，为后续接口调用铺底。
 * 具体的拦截器逻辑（鉴权注入、统一错误处理、数据解包等）
 * 待进入后端接口联调阶段再补充实现。
 */
const request: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器（占位：待接口开发阶段实现鉴权 / 通用头注入）
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // TODO: 注入 token / 租户标识等通用请求头
    return config
  },
  (error) => Promise.reject(error),
)

// 响应拦截器（占位：待接口开发阶段实现统一错误处理 / 数据解包）
request.interceptors.response.use(
  (response: AxiosResponse) => {
    // TODO: 统一处理后端返回结构（如 { code, message, data }）
    return response
  },
  (error) => {
    // TODO: 统一错误提示（如 401 跳登录、500 全局提示）
    return Promise.reject(error)
  },
)

export default request
