import React, { useEffect } from 'react'
import { message } from 'antd'
import './index.css'

// 提示消息全局缓存，避免在极短时间内（例如 1.5 秒）弹出相同的消息
const messageCache = new Set<string>()

const showFeedback = (type: 'success' | 'error' | 'warning', text: string) => {
  if (!text) return
  const cacheKey = `${type}_${text}`
  
  if (messageCache.has(cacheKey)) {
    return
  }
  
  messageCache.add(cacheKey)
  setTimeout(() => {
    messageCache.delete(cacheKey)
  }, 1500)
  
  if (type === 'success') {
    message.success(text)
  } else if (type === 'warning') {
    message.warning(text)
  } else {
    message.error(text)
  }
}

interface FeedbackProps {
  type?: 'success' | 'error' | 'warning' | null
  content?: string
}

/**
 * 统一的后端调用给用户反馈的反馈组件
 * 
 * 既支持声明式组件使用：
 * <Feedback type="success" content="基础数据加载成功" />
 * 
 * 也支持函数式（静态方法）调用：
 * Feedback.success("提交成功")
 * Feedback.warning("提示内容")
 * Feedback.error("获取数据失败")
 */
export const Feedback: React.FC<FeedbackProps> & {
  success: (content: string) => void
  warning: (content: string) => void
  error: (content: string) => void
  handle: (resOrError: any, defaultSuccessMsg?: string, defaultErrorMsg?: string) => void
} = ({ type, content }) => {
  useEffect(() => {
    if (type && content) {
      showFeedback(type, content)
    }
  }, [type, content])

  return <div className="feedback-placeholder" />
}

Feedback.success = (content: string) => {
  showFeedback('success', content)
}

Feedback.warning = (content: string) => {
  showFeedback('warning', content)
}

Feedback.error = (content: string) => {
  showFeedback('error', content)
}

Feedback.handle = (resOrError: any, defaultSuccessMsg?: string, defaultErrorMsg?: string) => {
  if (!resOrError) return

  // 判断是否是 Error 实例或包含错误特征的对象
  const isError = resOrError instanceof Error || 
                  resOrError.isAxiosError || 
                  resOrError.status === 'error' ||
                  (resOrError.code !== undefined && resOrError.code !== 200 && resOrError.code !== '200' && resOrError.code !== 'success')

  if (isError) {
    let errorMsg = defaultErrorMsg || '操作失败'
    if (resOrError instanceof Error) {
      errorMsg = resOrError.message
    } else if (resOrError.response?.data?.message) {
      errorMsg = resOrError.response.data.message
    } else if (resOrError.message) {
      errorMsg = resOrError.message
    } else if (typeof resOrError === 'string') {
      errorMsg = resOrError
    }
    showFeedback('error', errorMsg)
  } else {
    // 成功处理分支：识别请求方式与文案特征
    // 1. 获取请求 Method（从直接 config 或 axios 响应解包附加的 _config 中获取）
    const config = resOrError?.config || resOrError?._config
    const method = config?.method?.toLowerCase()

    // 2. 拼接成功提示文案
    let successMsg = ''
    if (resOrError.message) {
      successMsg = resOrError.message
    } else if (resOrError.msg) {
      successMsg = resOrError.msg
    } else if (typeof resOrError === 'string') {
      successMsg = resOrError
    } else if (defaultSuccessMsg) {
      successMsg = defaultSuccessMsg
    }

    // 3. 判断是否为纯查询/读取类操作：
    // - HTTP 请求方式为 GET
    // - 或者消息文案中包含常见的查询/读取关键字（查询、获取、加载、查看、列表等）
    const isQueryOperation = method === 'get'

    if (isQueryOperation) {
      return // 纯查询/加载数据操作，成功时静默，不弹出反馈提示
    }

    if (successMsg) {
      showFeedback('success', successMsg)
    }
  }
}

export default Feedback
