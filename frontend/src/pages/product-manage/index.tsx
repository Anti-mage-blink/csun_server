import React from 'react';
import { Card } from 'antd';
import { NativeDataTable } from '../../components/NativeDataTable';

const ProductManage: React.FC = () => {
  // 显示字段数组
  const showFields = [
    'main_category',
    'product_name',
    'spec',
    'high_threshold',
    'low_threshold',
    'big_batch_price',
    'middle_batch_price',
    'small_batch_price',
    'floor_price',
  ];

  // 可修改字段数组
  const editableFields = [
    'main_category',
    'product_name',
    'spec',
    'high_threshold',
    'low_threshold',
    'big_batch_price',
    'middle_batch_price',
    'small_batch_price',
    'floor_price',
  ];

  // 必填字段数组
  const requiredFields = [
    'main_category',
    'product_name',
    'high_threshold',
    'low_threshold',
    'big_batch_price',
    'middle_batch_price',
    'small_batch_price',
    'floor_price',
  ];

  // 字段 Label 对应字典
  const fieldLabelMap = {
    main_category: '产品分类',
    product_name: '产品名称',
    spec: '规格',
    high_threshold: '高批量阈值',
    low_threshold: '低批量阈值',
    big_batch_price: '大批量价格',
    middle_batch_price: '中批量价格',
    small_batch_price: '样品/小单价格',
    floor_price: '底线价',
  };

  // 哪些字段是选择类型（其余为填写类型），选项对应什么数据表的什么字段
  const selectFieldsMap = {
    main_category: {
      targetTable: 'quote_manage.main_category',
      targetField: 'main_category_name',
      valueField: 'main_category_name',
      labelField: 'main_category_name',
    },
  };

  // 字段输入控件渲染类型
  const fieldTypesMap: Record<string, 'text' | 'number' | 'digit' | 'date' | 'textarea'> = {
    high_threshold: 'number',
    low_threshold: 'number',
    big_batch_price: 'number',
    middle_batch_price: 'number',
    small_batch_price: 'number',
    floor_price: 'number',
  };

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <NativeDataTable
          tableStr="quote_manage.product_spec"
          tableDisplayName="产品数据表"
          relationTables={['quote_manage.main_category']}
          showFields={showFields}
          editableFields={editableFields}
          requiredFields={requiredFields}
          fieldLabelMap={fieldLabelMap}
          selectFieldsMap={selectFieldsMap}
          fieldTypesMap={fieldTypesMap}
        />
      </Card>
    </div>
  );
};

export default ProductManage;
