import React, { useState, useRef } from 'react'
import {
  Card,
  Typography,
  Tag,
  Button,
  Space,
  Alert,
  List,
  message,
  Divider,
  Input,
  Tooltip,
  Badge,
} from 'antd'
import {
  UploadOutlined,
  DownloadOutlined,
  DeleteOutlined,
  EyeOutlined,
  FileOutlined,
  SendOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import request from '@/api/request'
import { parseFilenameFromPath, downloadCosFile } from '@/utils/file'

const { Title, Paragraph, Text } = Typography

export { parseFilenameFromPath, downloadCosFile }

// 1. 临时待上传附件类型（保存在前端内存中）
interface TempAttachment {
  id: string
  file: File
  objectUrl: string // 本地 Blob 预览链接
  name: string
  size: number
}

// 2. COS 上传返回结果类型
interface UploadedResult {
  key: string
  filename: string
  size: number
  download_url: string
}

const MAX_FILE_SIZE = 16 * 1024 * 1024 // 16MB 限制

// 格式化文件大小
const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * 为本地 File 对象生成带有 charset=utf-8 编码声明的临时 Blob URL
 * 解决 Windows 环境下 Markdown (.md)、文本文件 (.txt) 缺少字符集描述导致浏览器在新标签页渲染为 GBK/ANSI 乱码的问题
 */
const createTextBlobUrl = (file: File): string => {
  let mimeType = file.type
  const fileName = file.name.toLowerCase()

  const isText =
    mimeType.startsWith('text/') ||
    fileName.endsWith('.md') ||
    fileName.endsWith('.txt') ||
    fileName.endsWith('.json') ||
    fileName.endsWith('.csv') ||
    fileName.endsWith('.xml') ||
    fileName.endsWith('.html') ||
    fileName.endsWith('.js') ||
    fileName.endsWith('.ts')

  if (isText) {
    if (!mimeType || mimeType === 'text/plain') {
      if (fileName.endsWith('.md')) mimeType = 'text/markdown'
      else if (fileName.endsWith('.json')) mimeType = 'application/json'
      else mimeType = 'text/plain'
    }
    if (!mimeType.includes('charset')) {
      mimeType = `${mimeType};charset=utf-8`
    }
    const blob = new Blob([file], { type: mimeType })
    return URL.createObjectURL(blob)
  }

  return URL.createObjectURL(file)
}

const TestPage: React.FC = () => {
  // --- 状态1：上传处 - 暂存的本地临时附件列表 ---
  const [tempAttachments, setTempAttachments] = useState<TempAttachment[]>([])
  const [submitting, setSubmitting] = useState<boolean>(false)

  // --- 状态2：存储与查看处 - 数据库字段 attachment_path_array 的当前值 ---
  const [attachmentPathArray, setAttachmentPathArray] = useState<string[]>([])

  // 手动测试使用的 JSON 输入串
  const [jsonInput, setJsonInput] = useState<string>(JSON.stringify([], null, 2))

  const fileInputRef = useRef<HTMLInputElement | null>(null)

  // ==================== 1. 上传处逻辑 ====================

  // 点击【上传附件/添加附件】按钮
  const handleSelectFilesClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.value = '' // 清空历史选择，保证重复选相同文件能正常触发
      fileInputRef.current.click()
    }
  }

  // 选择本地文件后的回调（处理多选 + 16MB 限制校验 + 本地暂存）
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (!files || files.length === 0) return

    const newTemps: TempAttachment[] = []
    let oversizeCount = 0

    Array.from(files).forEach((file) => {
      // 校验大小：必须 <= 16MB
      if (file.size > MAX_FILE_SIZE) {
        oversizeCount++
        return
      }

      // 校验通过，创建本地 Blob 临时访问链接（带 utf-8 编码声明，防止文本乱码）
      const objectUrl = createTextBlobUrl(file)
      newTemps.push({
        id: `${Date.now()}_${Math.random().toString(36).substring(2, 9)}`,
        file,
        objectUrl,
        name: file.name,
        size: file.size,
      })
    })

    if (oversizeCount > 0) {
      message.error(`${oversizeCount} 个文件大小超过 16MB 限制被跳过！`)
    }

    if (newTemps.length > 0) {
      setTempAttachments((prev) => [...prev, ...newTemps])
      message.success(`成功选择并暂存 ${newTemps.length} 个文件`)
    }
  }

  // 打开本地暂存的临时文件预览/下载
  const handleOpenTempFile = (item: TempAttachment) => {
    window.open(item.objectUrl, '_blank')
  }

  // 移除单个暂存文件
  const handleRemoveTempFile = (id: string) => {
    setTempAttachments((prev) => {
      const target = prev.find((item) => item.id === id)
      if (target) {
        URL.revokeObjectURL(target.objectUrl) // 释放内存
      }
      return prev.filter((item) => item.id !== id)
    })
    message.info('已从临时暂存列表中移除该文件')
  }

  // 清空所有暂存文件
  const handleClearAllTemp = () => {
    tempAttachments.forEach((item) => URL.revokeObjectURL(item.objectUrl))
    setTempAttachments([])
    message.info('已清空全部暂存文件')
  }

  // 模拟“提交报价单”：将所有临时暂存的文件批量/依次上传至 COS，生成相对路径数组
  const handleSubmitQuote = async () => {
    if (tempAttachments.length === 0) {
      message.warning('请先添加并暂存至少一个附件文件！')
      return
    }

    setSubmitting(true)
    const uploadedKeys: string[] = []

    try {
      for (const item of tempAttachments) {
        const formData = new FormData()
        formData.append('file', item.file)

        const res = await request.post('/cos/upload', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        })

        if (res.data && res.data.data && res.data.data.key) {
          const result: UploadedResult = res.data.data
          uploadedKeys.push(result.key)
        } else {
          message.error(`文件 [${item.name}] 上传失败：响应格式异常`)
        }
      }

      if (uploadedKeys.length > 0) {
        message.success(`成功上传 ${uploadedKeys.length} 个附件至 COS！并更新报价单 attachment_path_array 字段`)
        
        // 更新“查看处”的路径数组字段
        setAttachmentPathArray((prev) => {
          const newArray = [...prev, ...uploadedKeys]
          setJsonInput(JSON.stringify(newArray, null, 2))
          return newArray
        })

        // 清空暂存列表
        handleClearAllTemp()
      }
    } catch (err: any) {
      const errMsg = err.response?.data?.message || err.message || '上传过程发生错误'
      message.error(`提交报价单附件失败: ${errMsg}`)
    } finally {
      setSubmitting(false)
    }
  }

  // ==================== 2. 查看与下载处逻辑 ====================

  // 点击查看处的文件名或下载按钮，触发 COS 接口文件下载
  const handleDownloadCosFile = (pathKey: string) => {
    downloadCosFile(pathKey)
  }

  // 应用手动修改的 JSON 测试路径数组
  const handleApplyJsonInput = () => {
    try {
      const parsed = JSON.parse(jsonInput)
      if (Array.isArray(parsed) && parsed.every((item) => typeof item === 'string')) {
        setAttachmentPathArray(parsed)
        message.success('已应用新的 attachment_path_array 测试数据')
      } else {
        message.error('输入的 JSON 格式不正确，必须为字符串数组，例如 ["path1", "path2"]')
      }
    } catch (e) {
      message.error('解析 JSON 语法失败，请检查输入格式')
    }
  }

  return (
    <div style={{ padding: 24, maxWidth: 1100, margin: '0 auto' }}>
      {/* 隐藏的文件选择器，支持多选 multiple */}
      <input
        type="file"
        ref={fileInputRef}
        style={{ display: 'none' }}
        multiple
        onChange={handleFileChange}
      />

      {/* 头部标题卡片 */}
      <Card bordered={false} style={{ marginBottom: 24, borderRadius: 8, boxShadow: '0 2px 8px rgba(0,0,0,0.06)' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <Title level={3} style={{ marginBottom: 4 }}>
              报价单附件功能测试 (多文件暂存 + 云存储上传 + 路径解析下载)
            </Title>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              测试字段: <Text code>attachment_path_array</Text> (JSON 字符串数组) | COS Bucket: <Text code>csun-server-1444192538</Text>
            </Paragraph>
          </div>
          <Tag color="purple" style={{ fontSize: 14, padding: '4px 12px' }}>
            测试专用
          </Tag>
        </div>
      </Card>

      {/* 规则说明 */}
      <Alert
        message="附件全流程业务说明"
        description={
          <div>
            <div>1. <strong>上传处（新建/编辑报价单）</strong>：点击【添加附件】弹出文件选择框（支持多选）。选择后进行 16MB 大小校验，校验通过后仅<strong>临时暂存在前端</strong>（未调用后端/COS上传）。此时前端生成临时链接，点击文件名可打开预览。</div>
            <div>2. <strong>提交报价单</strong>：点击【模拟提交报价单】时，将所有暂存的附件统一上传至腾讯云 COS，拿到返回的相对路径列表写入 <Text code>attachment_path_array</Text> 字段。</div>
            <div>3. <strong>查看处（审批/备案详情）</strong>：从 <Text code>attachment_path_array</Text> 字段读取相对路径列表，经过<strong>“掐头去尾”</strong>算法自动解析提炼出清晰的原始文件名，展示为包含扩展名图标的链接，点击即可下载。</div>
          </div>
        }
        type="info"
        showIcon
        style={{ marginBottom: 24, borderRadius: 8 }}
      />

      {/* 第一部分：上传处测试 */}
      <Card
        title={
          <Space>
            <UploadOutlined style={{ color: '#1890ff' }} />
            <span>【上传处测试】附件选择与前端临时暂存</span>
            <Badge count={tempAttachments.length} showZero overflowCount={99} color="#52c41a" />
          </Space>
        }
        bordered={false}
        style={{ marginBottom: 24, borderRadius: 8, boxShadow: '0 2px 8px rgba(0,0,0,0.06)' }}
        extra={
          tempAttachments.length > 0 && (
            <Button danger type="text" icon={<DeleteOutlined />} onClick={handleClearAllTemp}>
              清空暂存
            </Button>
          )
        }
      >
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <Space size="middle">
            <Button type="primary" ghost icon={<UploadOutlined />} onClick={handleSelectFilesClick}>
              添加附件 (支持多选)
            </Button>
            <Button
              type="primary"
              icon={<SendOutlined />}
              onClick={handleSubmitQuote}
              loading={submitting}
              disabled={tempAttachments.length === 0}
            >
              模拟提交报价单 (统一上传至 COS)
            </Button>
          </Space>

          {/* 暂存文件列表 */}
          {tempAttachments.length > 0 ? (
            <List
              header={
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Text strong>已暂存的本地附件（未真正上传）</Text>
                  <Tag color="orange">临时暂存中</Tag>
                </div>
              }
              bordered
              dataSource={tempAttachments}
              renderItem={(item) => (
                <List.Item
                  actions={[
                    <Button
                      type="link"
                      size="small"
                      icon={<EyeOutlined />}
                      onClick={() => handleOpenTempFile(item)}
                    >
                      点击打开/预览
                    </Button>,
                    <Button
                      type="text"
                      danger
                      size="small"
                      icon={<DeleteOutlined />}
                      onClick={() => handleRemoveTempFile(item.id)}
                    >
                      移除
                    </Button>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={<FileTextOutlined style={{ fontSize: 24, color: '#1890ff' }} />}
                    title={
                      <a
                        onClick={(e) => {
                          e.preventDefault()
                          handleOpenTempFile(item)
                        }}
                        style={{ fontWeight: 500 }}
                      >
                        {item.name}
                      </a>
                    }
                    description={
                      <Space split={<Divider type="vertical" />}>
                        <span>大小: {formatFileSize(item.size)}</span>
                        <Tag color="blue">本地临时链接已生成</Tag>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {item.file.type || '二进制/未知类型'}
                        </Text>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          ) : (
            <Alert
              message="暂未选择任何附件文件，点击上方【添加附件】选择本地电脑中的文件进行测试。"
              type="warning"
              showIcon
            />
          )}
        </Space>
      </Card>

      {/* 第二部分：查看与解析处测试 */}
      <Card
        title={
          <Space>
            <FileOutlined style={{ color: '#52c41a' }} />
            <span>【查看处测试】<Text code>attachment_path_array</Text> 解析显示与下载</span>
          </Space>
        }
        bordered={false}
        style={{ borderRadius: 8, boxShadow: '0 2px 8px rgba(0,0,0,0.06)' }}
      >
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          {/* JSON 输入调试与呈现 */}
          <Card size="small" type="inner" title="数据库存储数据 (attachment_path_array 字段)">
            <Paragraph style={{ marginBottom: 8 }}>
              当前相对路径列表JSON：
            </Paragraph>
            <Input.TextArea
              rows={3}
              value={jsonInput}
              onChange={(e) => setJsonInput(e.target.value)}
              style={{ fontFamily: 'monospace', marginBottom: 8 }}
            />
            <Button size="small" type="default" onClick={handleApplyJsonInput}>
              更新测试数据并重解析
            </Button>
          </Card>

          <Divider style={{ margin: '8px 0' }} />

          {/* 解析后的附件下载渲染列表 */}
          <div>
            <Title level={5} style={{ marginBottom: 16 }}>
              前端经过“掐头去尾”提炼后的附件下载列表：
            </Title>

            {attachmentPathArray.length > 0 ? (
              <List
                grid={{ gutter: 16, xs: 1, sm: 2, md: 2, lg: 3 }}
                dataSource={attachmentPathArray}
                renderItem={(pathKey) => {
                  const displayName = parseFilenameFromPath(pathKey)
                  const ext = displayName.includes('.') ? displayName.split('.').pop()?.toUpperCase() : 'FILE'

                  return (
                    <List.Item>
                      <Card
                        size="small"
                        hoverable
                        style={{ borderRadius: 6, border: '1px solid #e8e8e8' }}
                        actions={[
                          <Button
                            type="link"
                            icon={<DownloadOutlined />}
                            onClick={() => handleDownloadCosFile(pathKey)}
                          >
                            下载附件
                          </Button>,
                        ]}
                      >
                        <Card.Meta
                          avatar={
                            <Tag color="cyan" style={{ fontSize: 12, padding: '2px 6px', marginTop: 4 }}>
                              {ext}
                            </Tag>
                          }
                          title={
                            <Tooltip title={`完整 COS Key: ${pathKey}`}>
                              <a
                                onClick={(e) => {
                                  e.preventDefault()
                                  handleDownloadCosFile(pathKey)
                                }}
                                style={{
                                  fontSize: 14,
                                  fontWeight: 600,
                                  color: '#1890ff',
                                  wordBreak: 'break-all',
                                }}
                              >
                                {displayName}
                              </a>
                            </Tooltip>
                          }
                          description={
                            <div style={{ marginTop: 6 }}>
                              <Text type="secondary" style={{ fontSize: 12 }} ellipsis={{ tooltip: pathKey }}>
                                相对路径: {pathKey}
                              </Text>
                            </div>
                          }
                        />
                      </Card>
                    </List.Item>
                  )
                }}
              />
            ) : (
              <Alert message="attachment_path_array 字段为空，暂无可下载的附件" type="info" showIcon />
            )}
          </div>
        </Space>
      </Card>
    </div>
  )
}

export default TestPage
