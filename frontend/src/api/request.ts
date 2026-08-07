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

// 响应拦截器（统一注入请求配置上下文，供 Feedback 等辅助组件识别方法与路径）
request.interceptors.response.use(
  (response: AxiosResponse) => {
    if (response.data && typeof response.data === 'object') {
      try {
        Object.defineProperty(response.data, '_config', {
          value: response.config,
          writable: true,
          enumerable: false, // 设为不可枚举，避免影响 JSON 遍历或数据处理
        })
      } catch (e) {
        // 忽略定义失败的特殊对象
      }
    }
    return response
  },
  (error) => {
    return Promise.reject(error)
  },
)

export default request
