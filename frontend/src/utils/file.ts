/**
 * 工具函数：对 COS 相对路径进行“掐头去尾”，提炼出原始文件名（含后缀）
 * 例如：
 * "test_uploads/20260807_150405_设计图纸.pdf" -> "设计图纸.pdf"
 * "quote_attachments/20260807_120000_报价清单.xlsx" -> "报价清单.xlsx"
 */
export const parseFilenameFromPath = (path: string): string => {
  if (!path) return ''
  // 1. 去除 URL Query 参数（若有）
  const cleanPath = path.split('?')[0]
  // 2. 截取最后一个 '/' 之后的文件名部分（去头部的目录路径）
  let baseName = cleanPath.split('/').pop() || cleanPath
  // 3. 循环去除自动生成的 15 位时间戳前缀 YYYYMMDD_HHMMSS_ (8位日期_6位时间_)
  const timestampRegex = /^\d{8}_\d{6}_(.+)$/
  let match = baseName.match(timestampRegex)
  while (match && match[1]) {
    baseName = match[1]
    match = baseName.match(timestampRegex)
  }
  return baseName
}

/**
 * 触发 COS 文件的直接下载（通过隐藏 iframe 触发，无需弹窗新页面，解决跳转闪烁问题）
 * @param pathKey COS 中的文件相对路径 key
 */
export const downloadCosFile = (pathKey: string): void => {
  if (!pathKey) return
  const downloadUrl = `/api/cos/download?key=${encodeURIComponent(pathKey)}`
  
  // 创建隐藏的 iframe 触发浏览器原生文件下载，不影响当前页面且不会打开新窗口
  const iframe = document.createElement('iframe')
  iframe.style.display = 'none'
  iframe.src = downloadUrl
  document.body.appendChild(iframe)
  
  // 下载触发后延迟清理 iframe 元素
  setTimeout(() => {
    if (document.body.contains(iframe)) {
      document.body.removeChild(iframe)
    }
  }, 60000)
}
