import React, { useState, useRef } from 'react'
import { Card, Typography, Tag, Button, Space, Alert, Table, message } from 'antd'
import { UploadOutlined, CloudUploadOutlined, DownloadOutlined, DeleteOutlined, CheckCircleOutlined } from '@ant-design/icons'
import request from '@/api/request'

const { Title, Paragraph, Text } = Typography

interface UploadedResult {
  key: string
  filename: string
  size: number
  download_url: string
}

const MAX_FILE_SIZE = 16 * 1024 * 1024 // 16MB

const TestPage: React.FC = () => {
  // 本地选中的待上传文件（尚未真正上传）
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [uploading, setUploading] = useState<boolean>(false)
  // 已成功上传的结果列表
  const [uploadedList, setUploadedList] = useState<UploadedResult[]>([])

  const fileInputRef = useRef<HTMLInputElement | null>(null)

  // 1. 点击“添加文件”触发本地文件选择器
  const handleAddClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.value = '' // 清空历史选择，确保能重复选择相同文件
      fileInputRef.current.click()
    }
  }

  // 2. 处理文件选择并做 16MB 校验
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (file.size > MAX_FILE_SIZE) {
      message.error(`添加失败：文件大小 (${(file.size / (1024 * 1024)).toFixed(2)} MB) 超过 16MB 限制！`)
      return
    }

    setSelectedFile(file)
    message.success(`文件 [${file.name}] 已添加至待上传列表`)
  }

  // 3. 点击“上传至 COS”按钮，发起后端请求
  const handleUpload = async () => {
    if (!selectedFile) {
      message.warning('请先添加要上传的文件！')
      return
    }

    setUploading(true)
    const formData = new FormData()
    formData.append('file', selectedFile)

    try {
      const res = await request.post('/test/cos/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })

      if (res.data && res.data.data) {
        const result: UploadedResult = res.data.data
        message.success(res.data.message || '文件成功上传至 COS 对象存储！')
        setUploadedList((prev) => [result, ...prev])
        setSelectedFile(null) // 清空待上传状态
      } else {
        message.error('上传响应格式异常')
      }
    } catch (err: any) {
      const errMsg = err.response?.data?.message || err.message || '上传 COS 失败'
      message.error(`上传失败: ${errMsg}`)
    } finally {
      setUploading(false)
    }
  }

  // 4. 文件下载触发
  const handleDownload = (record: UploadedResult) => {
    const downloadUrl = `/api/test/cos/download?key=${encodeURIComponent(record.key)}`
    window.open(downloadUrl, '_blank')
  }

  // 移除待上传文件
  const handleRemoveSelected = () => {
    setSelectedFile(null)
    message.info('已移除待上传文件')
  }

  // 格式化字节数
  const formatSize = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // 已上传结果表格列定义
  const columns = [
    {
      title: '原始文件名',
      dataIndex: 'filename',
      key: 'filename',
      render: (text: string) => <Text strong>{text}</Text>,
    },
    {
      title: 'COS 相对路径 (Key)',
      dataIndex: 'key',
      key: 'key',
      render: (text: string) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: '文件大小',
      dataIndex: 'size',
      key: 'size',
      render: (size: number) => formatSize(size),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: UploadedResult) => (
        <Button
          type="primary"
          ghost
          size="small"
          icon={<DownloadOutlined />}
          onClick={() => handleDownload(record)}
        >
          下载文件
        </Button>
      ),
    },
  ]

  return (
    <div style={{ padding: 24, maxWidth: 1000, margin: '0 auto' }}>
      <Card bordered={false} style={{ marginBottom: 24, borderRadius: 8 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <Title level={3} style={{ marginBottom: 4 }}>
              腾讯云 COS 对象存储测试 <Tag color="gold">上帝专属</Tag>
            </Title>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              Bucket: <Text code>csun-server-1444192538</Text> | Region: <Text code>ap-chengdu</Text>
            </Paragraph>
          </div>
        </div>
      </Card>

      <Card title="1. 选择并上传文件" bordered={false} style={{ marginBottom: 24, borderRadius: 8 }}>
        <input
          type="file"
          ref={fileInputRef}
          style={{ display: 'none' }}
          onChange={handleFileChange}
        />

        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <Alert
            message="使用规则说明"
            description="点击【添加文件】弹出系统文件管理器选择本地文件。校验文件大小需在 16MB 以内（超过 16MB 将提示添加失败）。添加时并不立即上传，选好文件后点击【上传至 COS】触发接口上传至存储桶并获取 COS 相对路径。"
            type="info"
            showIcon
          />

          <Space size="middle">
            <Button type="default" icon={<UploadOutlined />} onClick={handleAddClick}>
              添加文件
            </Button>
            <Button
              type="primary"
              icon={<CloudUploadOutlined />}
              onClick={handleUpload}
              loading={uploading}
              disabled={!selectedFile}
            >
              上传至 COS
            </Button>
          </Space>

          {/* 待上传文件展示卡片 */}
          {selectedFile && (
            <Card
              size="small"
              type="inner"
              title={
                <span>
                  <CheckCircleOutlined style={{ color: '#52c41a', marginRight: 8 }} />
                  待上传文件预览 (大小校验通过: {'<='} 16MB)
                </span>
              }
              extra={
                <Button
                  type="text"
                  danger
                  icon={<DeleteOutlined />}
                  onClick={handleRemoveSelected}
                >
                  移除
                </Button>
              }
              style={{ background: '#fafafa' }}
            >
              <Paragraph style={{ margin: 0 }}>
                <Text strong>文件名：</Text>{selectedFile.name}
              </Paragraph>
              <Paragraph style={{ margin: '4px 0 0 0' }}>
                <Text strong>文件大小：</Text>{formatSize(selectedFile.size)} ({selectedFile.size} 字节)
              </Paragraph>
              <Paragraph style={{ margin: '4px 0 0 0' }}>
                <Text strong>文件类型：</Text>{selectedFile.type || '未知/二进制类型'}
              </Paragraph>
            </Card>
          )}
        </Space>
      </Card>

      <Card title="2. COS 对象存储返回与下载结果" bordered={false} style={{ borderRadius: 8 }}>
        <Table
          dataSource={uploadedList}
          columns={columns}
          rowKey="key"
          pagination={{ pageSize: 5 }}
          locale={{ emptyText: '暂无已上传文件，请在上方添加并上传' }}
        />
      </Card>
    </div>
  )
}

export default TestPage
