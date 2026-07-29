import React from 'react'
import { Table, Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import './index.css'

/**
 * 下拉选择类型字段的配置
 */
export interface SelectFieldConfig {
  /** 本表中的字段名 */
  field: string
  /** 关联的目标数据库名（可选，不填则与主表相同） */
  targetDb?: string
  /** 关联的目标数据表名 */
  targetTable: string
  /** 选项值对应目标表的字段名（默认 id 或同名字段） */
  valueField?: string
  /** 选项显示文本对应目标表的字段名（默认 name / title 或同名字段） */
  labelField?: string
}

/**
 * NativeDataTable 通用原生表格组件参数接口
 */
export interface NativeDataTableProps {
  /** 数据库名 (例如: quote_manage) */
  dbName: string
  /** 表名 (例如: product_spec) */
  tableName: string
  /** 关联数据表列表: 同目标表做整表查询 */
  relatedTables?: string[]
  /** 显示哪些字段 (数组) */
  displayFields: string[]
  /** 可修改字段 (数组) */
  editableFields?: string[]
  /** 哪些字段为选择框（其余为输入框），及对应的目标数据表映射配置 */
  selectFields?: SelectFieldConfig[]
  /** 必填字段 (用于新增校验) */
  requiredFields?: string[]
  
  /** 字段中文名称映射，如 { product_name: '产品名称', spec: '规格型号' } */
  fieldLabels?: Record<string, string>
  
  /** 表格数据源 */
  dataSource?: Record<string, any>[]
  /** 加载中状态 */
  loading?: boolean
  /** 行唯一标识字段，默认 'id' */
  rowKey?: string
}

/**
 * NativeDataTable 原生数据表格组件
 * 
 * 用于通用数据表/字典表的展示与数据管理（CRUD）。
 */
const NativeDataTable: React.FC<NativeDataTableProps> = ({
  dbName,
  tableName,
  displayFields,
  requiredFields = [],
  fieldLabels = {},
  dataSource = [],
  loading = false,
  rowKey = 'id'
}) => {
  // 根据 displayFields 自动动态生成列配置
  const columns: ColumnsType<Record<string, any>> = displayFields.map((field) => {
    // 优先从 fieldLabels 读取中文别名，若无则显示原字段名
    const title = fieldLabels[field] || field
    // const isEditable = editableFields.includes(field)
    const isRequired = requiredFields.includes(field)
    // const isSelect = selectFields.some((item) => item.field === field)

    return {
      title: (
        <span>
          {title}
          {isRequired && <span style={{ color: '#ff4d4f', marginLeft: 4 }}>*</span>}
        </span>
      ),
      dataIndex: field,
      key: field,
      // 支持按此字段进行基础排序
      sorter: (a, b) => {
        const valA = a[field]
        const valB = b[field]
        if (typeof valA === 'number' && typeof valB === 'number') {
          return valA - valB
        }
        return String(valA ?? '').localeCompare(String(valB ?? ''))
      },
      render: (value: any) => {
        if (value === null || value === undefined) {
          return <span style={{ color: '#ccc' }}>-</span>
        }
        if (typeof value === 'boolean') {
          return <Tag color={value ? 'green' : 'red'}>{value ? '是' : '否'}</Tag>
        }
        return String(value)
      }
    }
  })

  return (
    <div className="native-data-table-container">
      <div className="native-data-table-header">
        <div className="native-data-table-title">
          <span>{tableName}</span>
          <span className="native-data-table-meta">
            ({dbName}.{tableName})
          </span>
        </div>
      </div>

      <Table
        rowKey={rowKey}
        columns={columns}
        dataSource={dataSource}
        loading={loading}
        pagination={{
          defaultPageSize: 10,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`
        }}
        bordered
        size="middle"
      />
    </div>
  )
}

export default NativeDataTable
