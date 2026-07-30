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
  inputPlaceholder?: string; // 新增填写模式下的输入框占位提示
  style?: React.CSSProperties; // 自定义样式
  className?: string;        // 自定义类名
  disableAddNew?: boolean;   // 是否完全禁用新增填写功能 (例如产品名称只允许从下拉中选择)
  disabled?: boolean;        // 是否禁用选择与新增填写
  forceNewMode?: boolean;    // 外部强制指定是否为“新增填写”模式
}

export const DropdownMenu: React.FC<DropdownMenuProps> = ({
  records = [],
  displayField,
  valueField,
  value = '',
  onChange,
  placeholder = '请选择',
  inputPlaceholder,
  style,
  className,
  disableAddNew = false,
  disabled = false,
  forceNewMode,
}) => {
  const actualValueField = valueField || displayField;

  // 是否处于“新增填写”模式
  const [isNewMode, setIsNewMode] = useState(false);
  // 输入框中的文字值
  const [inputValue, setInputValue] = useState('');
  // 选择框选中的值
  const [selectValue, setSelectValue] = useState<string | undefined>(undefined);

  // 记录上一次的 forceNewMode 状态
  const prevForceNewModeRef = React.useRef<boolean | undefined>(forceNewMode);

  // 根据 records 列表构造 antd 的 options，自动过滤掉 displayField 或 valueField 为空的无效记录
  const options = records
    .filter((record) => {
      if (!record) return false;
      const label = record[displayField];
      const val = record[actualValueField];
      const labelStr = label !== null && label !== undefined ? String(label).trim() : '';
      const valStr = val !== null && val !== undefined ? String(val).trim() : '';
      return labelStr !== '' && valStr !== '';
    })
    .map((record) => {
      const val = String(record[actualValueField] || '');
      const label = String(record[displayField] || '');
      return {
        label,
        value: val,
        record,
      };
    });

  // 在最后追加一个“新增填写”的选项
  const selectOptions = disableAddNew
    ? options
    : [
        ...options,
        { label: '➕ 新增填写', value: '__ADD_NEW__', record: null },
      ];

  // 同步外部传入的 value 及 forceNewMode 属性
  useEffect(() => {
    // 1. 如果外部强制开启新增模式
    if (forceNewMode === true) {
      setIsNewMode(true);
      setInputValue(value);
      setSelectValue(undefined);
      prevForceNewModeRef.current = forceNewMode;
      return;
    }

    // 2. 如果 forceNewMode 从 true 变为 false 或 undefined，取消强制新增，切回选择模式
    if (prevForceNewModeRef.current === true) {
      setIsNewMode(false);
      setInputValue('');
      setSelectValue(undefined);
      prevForceNewModeRef.current = forceNewMode;
      return;
    }

    prevForceNewModeRef.current = forceNewMode;

    // 3. 检查当前 value 是否存在于已有 records 中
    const exists = options.some((opt) => opt.value === value && value !== '');

    if (exists) {
      setIsNewMode(false);
      setSelectValue(value);
      setInputValue('');
    } else if (value !== '') {
      if (!disableAddNew) {
        setIsNewMode(true);
        setInputValue(value);
        setSelectValue(undefined);
      }
    } else {
      // value === '' 时：
      // 如果已处于新增模式（例如刚点击了“➕ 新增填写”导致 onChange('')，或在输入框中清空），保持新增模式并更新 input 框
      // 如果处于选择模式，维持选择模式，清除下拉选框选中项
      if (isNewMode) {
        setInputValue('');
      } else {
        setSelectValue(undefined);
      }
    }
  }, [value, records, actualValueField, displayField, disableAddNew, forceNewMode]);

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

  // 智能计算 Input 的 placeholder
  const computedInputPlaceholder =
    inputPlaceholder ||
    `请输入${placeholder.replace(/^(请选择|选择)/, '').replace(/或新增填写$/, '') || '内容'}`;

  if (isNewMode) {
    return (
      <Space.Compact style={style} className={`dropdown-menu-container ${className || ''}`}>
        <Input
          placeholder={computedInputPlaceholder}
          value={inputValue}
          onChange={handleInputChange}
          allowClear
          disabled={disabled}
        />
        <Button
          type="primary"
          icon={<RollbackOutlined />}
          onClick={handleRollback}
          title="返回下拉选择"
          disabled={disabled}
        />
      </Space.Compact>
    );
  }

  return (
    <Select
      showSearch
      disabled={disabled}
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
