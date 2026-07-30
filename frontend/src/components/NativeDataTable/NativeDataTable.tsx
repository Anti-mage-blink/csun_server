import React, { useState, useMemo, useRef } from 'react';
import {
  ProTable,
  ModalForm,
  ProFormText,
  ProFormSelect,
  ProFormDigit,
  ProFormDatePicker,
  ProFormTextArea,
  type ActionType,
  type ProColumns,
} from '@ant-design/pro-components';
import { Button, Input, Popconfirm, Space, Tag } from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  TableOutlined,
  CalendarOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import {
  fetchDataTableData,
  createDataTableRecord,
  updateDataTableRecord,
  deleteDataTableRecord,
} from '../../api/dataTable';
import Feedback from '../Feedback';
import './NativeDataTable.css';

/**
 * 支持的字段输入渲染类型
 */
export type FieldType = 'text' | 'number' | 'digit' | 'date' | 'textarea' | 'select';

/**
 * 关联选择字段配置，可为 "数据库名.数据表名" 字符串，或详细对象配置
 */
export type SelectRelationConfig =
  | string
  | {
      /** 关联的目标数据库名与数据表名，格式为 "数据库名.数据表名" 或 "数据表名" */
      targetTable: string;
      /** 关联数据表中的对应字段名 (如 "main_category_name") */
      targetField?: string;
      /** 关联数据表中作为选中的 value 字段名 (默认同 targetField) */
      valueField?: string;
      /** 关联数据表中作为显示的 label 字段名 (默认同 targetField) */
      labelField?: string;
    };

/**
 * 标准化的关联选择字段配置对象
 */
export interface ParsedSelectConfig {
  targetTable: string;
  valKey: string;
  labelKey: string;
}

/**
 * 解析各种格式的 SelectRelationConfig 为统一的标准对象结构
 */
export const parseSelectConfig = (
  config?: SelectRelationConfig
): ParsedSelectConfig => {
  if (!config) {
    return { targetTable: '', valKey: '', labelKey: '' };
  }

  if (typeof config === 'string') {
    const parts = config.trim().split('.');
    if (parts.length === 3) {
      return {
        targetTable: `${parts[0]}.${parts[1]}`,
        valKey: parts[2],
        labelKey: parts[2],
      };
    }
    return {
      targetTable: config.trim(),
      valKey: '',
      labelKey: '',
    };
  }

  const targetTable = config.targetTable ? config.targetTable.trim() : '';
  const fieldKey = config.targetField || config.valueField || config.labelField || '';
  const valKey = config.valueField || fieldKey;
  const labelKey = config.labelField || fieldKey;

  return { targetTable, valKey, labelKey };
};

/**
 * 从关联表数据项中通用提取下拉框的 value 和 label，与具体业务数据表解除字段名硬编码耦合
 */
export const extractOptionValueLabel = (
  item: any,
  field: string,
  valKey?: string,
  labelKey?: string
): { value: any; label: string } => {
  if (!item || typeof item !== 'object') {
    return { value: item, label: String(item ?? '') };
  }

  // 1. 计算 value (指定了 valKey 则优先使用，没有则回退到通用机制)
  let value = valKey && item[valKey] !== undefined ? item[valKey] : undefined;
  if (value === undefined) {
    value = item[field] ?? item.id ?? item.value;
  }

  // 2. 计算 label (指定了 labelKey 则优先使用，没有则回退到通用标准名称字段)
  let label = labelKey && item[labelKey] !== undefined ? item[labelKey] : undefined;
  if (label === undefined) {
    label = item[field] ?? item.name ?? item.title ?? item.label ?? value;
  }

  return {
    value,
    label: String(label ?? ''),
  };
};

/**
 * NativeDataTable 通用数据表 CRUD 组件 Props
 */
export interface NativeDataTableProps {
  /** "数据库名.数据表名" 字符串 (如 "general.customer") */
  tableStr?: string;
  /** 核心数据表的显示名 (如 "产品数据表") */
  tableDisplayName?: string;
  /** 兼容旧版：数据库名 */
  dbName?: string;
  /** 兼容旧版：数据表名 */
  tableName?: string;
  /** 关联数据表列表，格式为 "数据库名.数据表名" 数组 (同目标表做整表查询) */
  relationTables?: string[];
  /** 显示哪些字段 (数组) */
  showFields: string[];
  /** 可修改字段 (数组)，用于新增/编辑表单 */
  editableFields: string[];
  /** 哪些字段选择 (其余为填写)，对应什么数据表，格式为 { [fieldName]: "数据库名.数据表名" | Config } */
  selectFieldsMap?: Record<string, SelectRelationConfig>;
  /** 字段输入渲染类型映射，格式为 { [fieldName]: FieldType }，用于弹窗表单控件区分 */
  fieldTypesMap?: Record<string, FieldType>;
  /** 新增组件参数：显示的数据表字段名和用于前端显示的 label-value 对应关系 */
  fieldLabelMap?: Record<string, string>;
  /** 关联字段选择框的选项映射 { [fieldName]: Array<{ label: string; value: any }> } */
  relationOptionsMap?: Record<string, Array<{ label: string; value: any }>>;
  /** 必填字段 (用于新增校验) */
  requiredFields?: string[];
  /** 该数据表全量字段列表，用于构建 1:1 结构体 (若不传则从 showFields + editableFields 推导) */
  allFields?: string[];
  /** 外部传入的数据列表 (纯前端展示/测试模式) */
  dataSource?: Record<string, any>[];
  /** 表头标题（可选，如果不传则优先使用 tableDisplayName） */
  title?: string;
  /** 供外部操作组件的 actionRef */
  actionRef?: React.Ref<ActionType>;
  /** 自定义数据加载回调 */
  onRequestData?: (params: { searchKeyword: string; sortField?: string; sortOrder?: string }) => Promise<Record<string, any>[]> | Record<string, any>[];
  /** 新增记录事件回调 */
  onCreate?: (record: Record<string, any>) => Promise<boolean | void> | boolean | void;
  /** 修改记录事件回调 */
  onUpdate?: (id: number | string, record: Record<string, any>) => Promise<boolean | void> | boolean | void;
  /** 删除记录事件回调 */
  onDelete?: (id: number | string) => Promise<boolean | void> | boolean | void;
  /** 成功提交/删除后的回调 */
  onDataChange?: () => void;
}

export const NativeDataTable: React.FC<NativeDataTableProps> = ({
  tableStr: rawTableStr,
  tableDisplayName,
  dbName,
  tableName,
  relationTables = [],
  showFields,
  editableFields,
  selectFieldsMap = {},
  fieldTypesMap = {},
  fieldLabelMap = {},
  relationOptionsMap = {},
  requiredFields = [],
  allFields,
  dataSource,
  title,
  actionRef: externalActionRef,
  onRequestData,
  onCreate,
  onUpdate,
  onDelete,
  onDataChange,
}) => {
  const internalActionRef = useRef<ActionType>();
  const actionRef = (externalActionRef as React.MutableRefObject<ActionType | undefined>) || internalActionRef;

  // 统一解析出的主表标识字符串 "数据库名.数据表名"
  const tableStr = useMemo(() => {
    if (rawTableStr && rawTableStr.trim()) return rawTableStr.trim();
    if (dbName && tableName) return `${dbName}.${tableName}`;
    if (tableName) return tableName;
    return 'data_table';
  }, [rawTableStr, dbName, tableName]);

  // 内部内存数据状态（仅当未使用后端接口且未传 dataSource 时维护本地列表）
  const [localData, setLocalData] = useState<Record<string, any>[]>(dataSource || []);

  // 存储从后端返回的关联表选项映射
  const [fetchedRelationOptionsMap, setFetchedRelationOptionsMap] = useState<
    Record<string, Array<{ label: string; value: any }>>
  >({});

  // 搜集所有的关联表标识字符串列表 (去除重复)
  const allRelationTables = useMemo(() => {
    const tableSet = new Set<string>();
    (relationTables || []).forEach((t) => t && tableSet.add(t.trim()));

    Object.values(selectFieldsMap).forEach((config) => {
      const parsed = parseSelectConfig(config);
      if (parsed.targetTable) {
        tableSet.add(parsed.targetTable);
      }
    });

    return Array.from(tableSet);
  }, [relationTables, selectFieldsMap]);

  // 最终的下拉选项映射 (优先使用外部传入的 relationOptionsMap，其次使用自动异步查出的选项)
  const finalRelationOptionsMap = useMemo(() => {
    return {
      ...fetchedRelationOptionsMap,
      ...relationOptionsMap,
    };
  }, [fetchedRelationOptionsMap, relationOptionsMap]);

  // 模糊搜索关键词
  const [searchKeyword, setSearchKeyword] = useState<string>('');

  // 弹窗状态管理
  const [modalOpen, setModalOpen] = useState<boolean>(false);
  const [editingRecord, setEditingRecord] = useState<Record<string, any> | null>(null);

  // 确定 1:1 结构体字段集合
  const fullFields = useMemo(() => {
    if (allFields && allFields.length > 0) {
      return Array.from(new Set(allFields));
    }
    return Array.from(new Set([...showFields, ...editableFields]));
  }, [allFields, showFields, editableFields]);

  // 构建 ProTable valueEnum 转换映射
  const valueEnumMap = useMemo(() => {
    const enumMap: Record<string, Record<string | number, { text: string }>> = {};
    Object.entries(finalRelationOptionsMap).forEach(([field, options]) => {
      const vEnum: Record<string | number, { text: string }> = {};
      options.forEach((opt) => {
        vEnum[opt.value] = { text: opt.label };
      });
      enumMap[field] = vEnum;
    });
    return enumMap;
  }, [finalRelationOptionsMap]);

  // 1. 构造 ProTable 列定义 columns，接入 fieldLabelMap 的对应关系与 fieldTypesMap 类型
  const columns: ProColumns<Record<string, any>>[] = useMemo(() => {
    const cols: ProColumns<Record<string, any>>[] = showFields.map((field, colIndex) => {
      const isSelect = Boolean(selectFieldsMap[field]);
      const valueEnum = valueEnumMap[field];
      const options = finalRelationOptionsMap[field];
      const displayTitle = fieldLabelMap[field] || field;
      const fieldType = fieldTypesMap[field] || 'text';

      let valueType: ProColumns<Record<string, any>>['valueType'] = isSelect ? 'select' : 'text';
      if (!isSelect) {
        if (fieldType === 'number' || fieldType === 'digit') valueType = 'digit';
        else if (fieldType === 'date') valueType = 'date';
        else if (fieldType === 'textarea') valueType = 'textarea';
      }

      return {
        title: displayTitle,
        dataIndex: field,
        key: field,
        // ProTable 原生列排序（升序 / 降序）
        sorter: true,
        valueType,
        valueEnum: isSelect && valueEnum ? valueEnum : undefined,
        fieldProps: isSelect ? { options } : undefined,
        render: (dom, record) => {
          const rawValue = record[field];

          // 1. 空值优雅渲染 (null, undefined, '')
          if (rawValue === null || rawValue === undefined || rawValue === '') {
            return <span className="ndt-empty-cell">—</span>;
          }

          // 2. 下拉/关联分类选择框 (Tag 呈现)
          if (isSelect) {
            let labelText: React.ReactNode = dom;
            if (valueEnum && valueEnum[rawValue]?.text) {
              labelText = valueEnum[rawValue].text;
            } else if (options) {
              const matchOpt = options.find(
                (opt) => opt.value === rawValue || String(opt.value) === String(rawValue)
              );
              if (matchOpt) labelText = matchOpt.label;
            }

            return <Tag className="ndt-select-tag">{labelText}</Tag>;
          }

          // 3. 数字 / 价格类型 (tabular-nums 财务级排版)
          if (fieldType === 'number' || fieldType === 'digit') {
            const isPrice = /price|cost|amount|fee|money/i.test(field);
            const num = Number(rawValue);

            if (!isNaN(num)) {
              const formattedNum = num.toLocaleString('zh-CN', {
                maximumFractionDigits: 4,
              });

              return (
                <span className={`ndt-number-cell ${isPrice ? 'ndt-price-cell' : ''}`}>
                  {isPrice && <span className="ndt-currency-symbol">￥</span>}
                  <span className="ndt-number-val">{formattedNum}</span>
                </span>
              );
            }
          }

          // 4. 日期类型 (附带微型日历图标)
          if (fieldType === 'date') {
            return (
              <span className="ndt-date-cell">
                <CalendarOutlined className="ndt-date-icon" />
                <span>{String(rawValue)}</span>
              </span>
            );
          }

          // 5. 文本/其他类型 (首列突出主标题感)
          const isFirstCol = colIndex === 0;
          return (
            <span className={`ndt-text-cell ${isFirstCol ? 'ndt-primary-text' : ''}`}>
              {dom}
            </span>
          );
        },
      };
    });

    // 追加操作列：修改与删除
    cols.push({
      title: '操作',
      valueType: 'option',
      key: 'option',
      width: 140,
      fixed: 'right',
      render: (_, record) => (
        <Space size={4}>
          <Button
            type="text"
            size="small"
            icon={<EditOutlined className="ndt-action-icon edit" />}
            className="ndt-action-btn edit"
            onClick={() => {
              setEditingRecord(record);
              setModalOpen(true);
            }}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除此条记录吗？"
            description="删除后将无法恢复"
            okText="确定"
            cancelText="取消"
            okButtonProps={{ danger: true, size: 'small' }}
            cancelButtonProps={{ size: 'small' }}
            onConfirm={() => handleDelete(record.id)}
          >
            <Button
              type="text"
              danger
              size="small"
              icon={<DeleteOutlined className="ndt-action-icon delete" />}
              className="ndt-action-btn delete"
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    });

    return cols;
  }, [showFields, selectFieldsMap, valueEnumMap, finalRelationOptionsMap, fieldLabelMap, fieldTypesMap]);

  // 2. 删除逻辑 (优先使用外部 onDelete 回调，默认调用专用 data_table 后端接口)
  const handleDelete = async (id: number | string) => {
    try {
      if (onDelete) {
        await onDelete(id);
      } else if (dataSource) {
        setLocalData((prev) => prev.filter((item) => item.id !== id));
        Feedback.success('记录已从前端列表中移除');
      } else {
        // 调用通用 data_table 删接口
        const res = await deleteDataTableRecord(tableStr, id);
        Feedback.handle(res, '记录删除成功', '删除记录失败');
      }
      actionRef.current?.reload();
      onDataChange?.();
    } catch (err: any) {
      Feedback.handle(err, undefined, '删除记录失败');
    }
  };

  // 3. 表单提交逻辑 (构建与数据表所有字段 1:1 的结构体，调用专用 data_table 增改接口)
  const handleFormFinish = async (formData: Record<string, any>) => {
    try {
      // 构建全字段 1:1 结构体 (没填则留 null)
      const recordStruct: Record<string, any> = {};

      fullFields.forEach((field) => {
        const val = formData[field];
        if (val !== undefined && val !== null && val !== '') {
          recordStruct[field] = val;
        } else if (editingRecord && editingRecord[field] !== undefined && editingRecord[field] !== null && editingRecord[field] !== '') {
          recordStruct[field] = editingRecord[field];
        } else {
          recordStruct[field] = null; // 没填留 null
        }
      });

      if (editingRecord && editingRecord.id !== undefined && editingRecord.id !== null) {
        // 修改操作
        if (onUpdate) {
          await onUpdate(editingRecord.id, recordStruct);
        } else if (dataSource) {
          setLocalData((prev) =>
            prev.map((item) => (item.id === editingRecord.id ? { ...item, ...recordStruct } : item))
          );
          Feedback.success('记录更新成功');
        } else {
          // 调用通用 data_table 改接口
          const res = await updateDataTableRecord(tableStr, editingRecord.id, recordStruct);
          Feedback.handle(res, '记录修改成功', '修改记录失败');
        }
      } else {
        // 新增操作
        if (onCreate) {
          await onCreate(recordStruct);
        } else if (dataSource) {
          const newRecord = { id: Date.now(), ...recordStruct };
          setLocalData((prev) => [newRecord, ...prev]);
          Feedback.success('新记录添加成功');
        } else {
          // 调用通用 data_table 增接口
          const res = await createDataTableRecord(tableStr, recordStruct);
          Feedback.handle(res, '记录新增成功', '新增记录失败');
        }
      }

      setModalOpen(false);
      actionRef.current?.reload();
      onDataChange?.();
      return true;
    } catch (err: any) {
      Feedback.handle(err, undefined, '保存记录失败');
      return false;
    }
  };

  return (
    <div className="native-data-table-wrapper">
      <ProTable<Record<string, any>>
        actionRef={actionRef}
        rowKey="id"
        headerTitle={
          <div className="native-data-table-title">
            <span className="native-data-table-title-icon">
              <TableOutlined />
            </span>
            <span>{title || tableDisplayName}</span>
          </div>
        }
        options={{ density: false }}
        columns={columns}
        scroll={{ x: 'max-content' }}
        // 使用原生 ProTable 的 request 回调进行数据展现、排序和模糊搜索过滤
        request={async (params, sort) => {
          const { current = 1, pageSize = 20 } = params;
          void current;
          void pageSize;

          let rawList: Record<string, any>[] = [];

          if (onRequestData) {
            const rawSortOrder = sort[Object.keys(sort)[0]];
            rawList = await onRequestData({
              searchKeyword,
              sortField: Object.keys(sort)[0],
              sortOrder: rawSortOrder ? String(rawSortOrder) : undefined,
            });
          } else if (dataSource) {
            rawList = dataSource;
          } else {
            // 补全对专用 data_table 查接口的调用：获取主表全量数据及所有关联数据表
            try {
              const res = await fetchDataTableData(tableStr, allRelationTables);
              Feedback.handle(res, undefined, '通过 data_table 接口加载数据失败');
              rawList = res.data || [];

              // 自动将后端返回的关联表 relations 转化为下拉框选项格式
              if (res.relations) {
                const newOptionsMap: Record<string, Array<{ label: string; value: any }>> = {};

                Object.entries(selectFieldsMap).forEach(([field, config]) => {
                  const { targetTable, valKey, labelKey } = parseSelectConfig(config);

                  if (targetTable && res.relations?.[targetTable]) {
                    const list = res.relations[targetTable];
                    const opts = list.map((item: any) =>
                      extractOptionValueLabel(item, field, valKey, labelKey)
                    );
                    newOptionsMap[field] = opts;
                  }
                });

                setFetchedRelationOptionsMap(newOptionsMap);
              }
            } catch (err: any) {
              Feedback.handle(err, undefined, '通过 data_table 接口加载数据失败');
              rawList = localData;
            }
          }

          // 1. 模糊搜索匹配 (遍历显示字段的值)
          let result = rawList;
          if (searchKeyword && searchKeyword.trim() !== '') {
            const kw = searchKeyword.trim().toLowerCase();
            result = result.filter((item) =>
              showFields.some((field) => {
                const val = item[field];
                if (val === null || val === undefined) return false;
                return String(val).toLowerCase().includes(kw);
              })
            );
          }

          // 2. 按某字段排序 (利用 ProTable 原生 sort 参数)
          const sortField = Object.keys(sort)[0];
          const sortOrder = sort[sortField]; // 'ascend' | 'descend'

          if (sortField && sortOrder) {
            result = [...result].sort((a, b) => {
              const valA = a[sortField];
              const valB = b[sortField];

              if (valA === valB) return 0;
              if (valA === null || valA === undefined) return 1;
              if (valB === null || valB === undefined) return -1;

              let cmp = 0;
              if (typeof valA === 'number' && typeof valB === 'number') {
                cmp = valA - valB;
              } else {
                cmp = String(valA).localeCompare(String(valB), 'zh-CN');
              }

              return sortOrder === 'ascend' ? cmp : -cmp;
            });
          }

          return {
            data: result,
            success: true,
            total: result.length,
          };
        }}
        // 关闭内置的多输入框 Search Form，在工具栏展现统一的上方模糊搜索框
        search={false}
        toolBarRender={() => [
          <Input
            key="fuzzy-search"
            className="native-data-table-search-input"
            placeholder="在表格字段中模糊搜索..."
            prefix={<SearchOutlined style={{ color: '#c9cdd4' }} />}
            allowClear
            value={searchKeyword}
            onChange={(e) => {
              const val = e.target.value;
              setSearchKeyword(val);
              if (val === '') {
                actionRef.current?.reload();
              }
            }}
            onPressEnter={() => {
              actionRef.current?.reload();
            }}
          />,
          <Button
            key="refresh-btn"
            className="native-data-table-refresh-btn"
            icon={<ReloadOutlined />}
            onClick={() => {
              actionRef.current?.reload();
            }}
            title="刷新表格"
          />,
          <Button
            key="create-btn"
            type="primary"
            icon={<PlusOutlined />}
            className="native-data-table-create-btn"
            onClick={() => {
              setEditingRecord(null);
              setModalOpen(true);
            }}
          >
            新增记录
          </Button>,
        ]}
      />

      {/* 动态增改 Modal 表单 */}
      <ModalForm
        title={editingRecord ? `编辑记录 (${tableDisplayName || tableStr})` : `新增记录 (${tableDisplayName || tableStr})`}
        open={modalOpen}
        initialValues={editingRecord || {}}
        modalProps={{
          onCancel: () => setModalOpen(false),
          destroyOnClose: true,
        }}
        onFinish={handleFormFinish}
      >
        {editableFields.map((field) => {
          const isSelect = Boolean(selectFieldsMap[field]);
          const isRequired = requiredFields.includes(field);
          const displayLabel = fieldLabelMap[field] || field;
          const fieldType = fieldTypesMap[field] || 'text';
          const rules = isRequired
            ? [{ required: true, message: `请${isSelect ? '选择' : '填写'} ${displayLabel}` }]
            : undefined;

          if (isSelect) {
            return (
              <ProFormSelect
                key={field}
                name={field}
                label={displayLabel}
                rules={rules}
                options={finalRelationOptionsMap[field] || []}
                placeholder={`请选择 ${displayLabel}`}
              />
            );
          }

          if (fieldType === 'number' || fieldType === 'digit') {
            return (
              <ProFormDigit
                key={field}
                name={field}
                label={displayLabel}
                rules={rules}
                placeholder={`请输入 ${displayLabel}`}
              />
            );
          }

          if (fieldType === 'date') {
            return (
              <ProFormDatePicker
                key={field}
                name={field}
                label={displayLabel}
                rules={rules}
                fieldProps={{ style: { width: '100%' } }}
                placeholder={`请选择 ${displayLabel}`}
              />
            );
          }

          if (fieldType === 'textarea') {
            return (
              <ProFormTextArea
                key={field}
                name={field}
                label={displayLabel}
                rules={rules}
                placeholder={`请输入 ${displayLabel}`}
              />
            );
          }

          return (
            <ProFormText
              key={field}
              name={field}
              label={displayLabel}
              rules={rules}
              placeholder={`请输入 ${displayLabel}`}
            />
          );
        })}
      </ModalForm>
    </div>
  );
};
