import React, { useState, useEffect } from 'react';
import { Select, Input, Button, Space } from 'antd';
import { RollbackOutlined } from '@ant-design/icons';
import './dropdownMenu.css';

export interface DropdownMenuProps {
  records?: any[];           // 某数据表的记录列表
  displayField: string;      // 指定显示的字段 (例如 'customer_name' 或 'name')
  valueField?: string;       // 值字段 (可选，如果未指定，默认使用 displayField)
  value?: string;            // 受控 value
  onChange?: (value: string, isNew: boolean, record?: any) => void; // 变更回调，当 isNew 为 true 时，表示为用户手动新增填写
  placeholder?: string;      // 占位提示
  style?: React.CSSProperties; // 自定义样式
  className?: string;        // 自定义类名
}

export const DropdownMenu: React.FC<DropdownMenuProps> = ({
  records = [],
  displayField,
  valueField,
  value = '',
  onChange,
  placeholder = '请选择',
  style,
  className,
}) => {
  const actualValueField = valueField || displayField;

  // 是否处于“新增填写”模式
  const [isNewMode, setIsNewMode] = useState(false);
  // 输入框中的文字值
  const [inputValue, setInputValue] = useState('');
  // 选择框选中的值
  const [selectValue, setSelectValue] = useState<string | undefined>(undefined);

  // 根据 records 列表构造 antd 的 options
  const options = records.map((record) => {
    const val = String(record[actualValueField] || '');
    const label = String(record[displayField] || '');
    return {
      label,
      value: val,
      record,
    };
  });

  // 在最后追加一个“新增填写”的选项
  const selectOptions = [
    ...options,
    { label: '➕ 新增填写', value: '__ADD_NEW__', record: null },
  ];

  // 同步外部传入的 value
  useEffect(() => {
    if (value === '') {
      // 如果值为空，重置状态
      setIsNewMode(false);
      setInputValue('');
      setSelectValue(undefined);
      return;
    }

    // 检查这个值是否在已有的 records 中存在
    const exists = options.some((opt) => opt.value === value);

    if (exists) {
      setIsNewMode(false);
      setSelectValue(value);
      setInputValue('');
    } else {
      // 外部传入的值不在已有列表中，说明可能处于“新增填写”态
      setIsNewMode(true);
      setInputValue(value);
      setSelectValue(undefined);
    }
  }, [value, records, actualValueField, displayField]);

  // 处理 Select 下拉菜单的选择
  const handleSelectChange = (val: string, option: any) => {
    if (val === '__ADD_NEW__') {
      setIsNewMode(true);
      setSelectValue(undefined);
      setInputValue('');
      if (onChange) {
        onChange('', true, null);
      }
    } else {
      setSelectValue(val);
      if (onChange) {
        onChange(val, false, option.record);
      }
    }
  };

  // 处理 Input 输入框的文字输入
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setInputValue(val);
    if (onChange) {
      onChange(val, true, null);
    }
  };

  // 点击“返回选择”按钮，切回下拉选择模式
  const handleRollback = () => {
    setIsNewMode(false);
    setInputValue('');
    setSelectValue(undefined);
    if (onChange) {
      onChange('', false, null);
    }
  };

  if (isNewMode) {
    return (
      <Space.Compact style={style} className={`dropdown-menu-container ${className || ''}`}>
        <Input
          placeholder={`请输入新增的${placeholder}`}
          value={inputValue}
          onChange={handleInputChange}
          allowClear
        />
        <Button
          type="primary"
          icon={<RollbackOutlined />}
          onClick={handleRollback}
          title="返回下拉选择"
        />
      </Space.Compact>
    );
  }

  return (
    <Select
      showSearch
      style={style}
      className={`dropdown-menu-container ${className || ''}`}
      placeholder={placeholder}
      value={selectValue}
      onChange={handleSelectChange}
      options={selectOptions}
      optionFilterProp="label"
      filterOption={(input, option) => {
        const labelStr = String(option?.label || '');
        return labelStr.toLowerCase().includes(input.toLowerCase());
      }}
    />
  );
};

export default DropdownMenu;
