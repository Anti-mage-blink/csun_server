import { useState } from 'react'
import { Typography, Button, List, Tag, message, Space, Card } from 'antd'
import { DatabaseOutlined, ReloadOutlined } from '@ant-design/icons'
import request from '@/api/request'
import styles from './index.module.css'

const { Title, Paragraph, Text } = Typography

interface TablesResponse {
  database: string
  tables: string[]
  count: number
  success: boolean
}

/**
 * 新建报价单模块（占位）
 * 当前提供一个"查询数据库表"按钮，用于验证 前端 -> 后端 -> 数据库 链路
 */
const QuoteCreate = () => {
  const [loading, setLoading] = useState(false)
  const [tables, setTables] = useState<string[]>([])
  const [database, setDatabase] = useState('')
  const [queried, setQueried] = useState(false)

  const handleQueryTables = async () => {
    setLoading(true)
    try {
      const res = await request.get<TablesResponse>('/tables')
      const data = res.data
      setTables(data.tables || [])
      setDatabase(data.database || '')
      setQueried(true)
      message.success(`查询成功，共 ${data.count} 张表`)
    } catch (err: any) {
      message.error(err?.message || '查询失败')
      setTables([])
      setQueried(true)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.container}>
      <Title level={4}>新建报价单</Title>
      <Paragraph type="secondary">模块开发中，待后续迭代实现。</Paragraph>

      <Card
        size="small"
        title={
          <Space>
            <DatabaseOutlined />
            <span>链路验证：查询数据库表</span>
          </Space>
        }
        style={{ maxWidth: 600, marginTop: 16 }}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Button
            type="primary"
            icon={<ReloadOutlined />}
            loading={loading}
            onClick={handleQueryTables}
          >
            查询 quote_manage 数据库表
          </Button>

          {queried && (
            <div>
              {database && (
                <Paragraph style={{ marginBottom: 8 }}>
                  <Text type="secondary">数据库：</Text>
                  <Tag color="blue">{database}</Tag>
                  <Text type="secondary">共 {tables.length} 张表</Text>
                </Paragraph>
              )}
              {tables.length > 0 ? (
                <List
                  size="small"
                  bordered
                  dataSource={tables}
                  renderItem={(name) => (
                    <List.Item>
                      <Tag color="geekblue">{name}</Tag>
                    </List.Item>
                  )}
                />
              ) : (
                <Text type="secondary">未查询到任何表</Text>
              )}
            </div>
          )}
        </Space>
      </Card>
    </div>
  )
}

export default QuoteCreate
