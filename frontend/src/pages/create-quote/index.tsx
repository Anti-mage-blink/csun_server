import React, { useState, useEffect } from 'react';
import {
  Form,
  Row,
  Col,
  Input,
  Typography,
  Card,
  Spin,
  Button,
  Table,
  Select,
  InputNumber,
  Popconfirm,
  Space
} from 'antd';
import { SaveOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { useAuth } from '@/AuthContext';
import DropdownMenu from '@/components/dropdownMenu';
import { enterCreateQuoteApi, submitQuoteApi, type CustomerRecord, type Quote, type SubmitQuotePayload } from '@/api/quote';
import Feedback from '@/components/Feedback';
import './index.css';

const { Title } = Typography;
const { Option } = Select;

// 扩展前端明细结构（继承自后端的骨架，并加入 UI 辅助属性）
interface FormQuoteItem {
  key: string;                    // 前端 Table 必须的唯一标识键
  product_category_id: number | null;
  product_category_name: string;
  product_name_id: number | null;
  product_name: string;
  product_spec_id: number | null;
  product_spec: string;
  price_catalog_version: string;
  order_batch_tier: string;      // 大批量, 中小批量, 样品/小单
  catalog_base_price: number;
  quote_float_rate: number;      // 前端浮动比例(如 5 代表 5%)
  quote_unit_price: number;
  quantity: number | null;
  total_amount: number;
  is_below_floor_price: boolean;

  // 联动及校验所需的宽表缓存字段
  high_threshold: number;
  low_threshold: number;
  big_batch_price: number;
  middle_small_batch_price: number;
  sample_small_order_price: number;
  quantityError?: string;        // 数量校验错误信息
}

const CreateQuote: React.FC = () => {
  const { user } = useAuth();
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  // 基础数据字典
  const [allCustomers, setAllCustomers] = useState<CustomerRecord[]>([]);
  const [uniqueCustomers, setUniqueCustomers] = useState<CustomerRecord[]>([]);
  const [filteredContacts, setFilteredContacts] = useState<CustomerRecord[]>([]);
  const [productSpecsList, setProductSpecsList] = useState<any[]>([]);
  const [isNewCustomer, setIsNewCustomer] = useState(false);

  // ================= 核心双向绑定状态 =================
  // 1. 报价单结构体双向状态 (直接绑定后端返回结构体，并实时更新)
  const [quote, setQuote] = useState<Quote | null>(null);

  // 2. 报价明细数组双向状态 (初始化为后端返回空明细的单元素数组)
  const [quoteItems, setQuoteItems] = useState<FormQuoteItem[]>([]);

  // 后端返回的空明细单原型，用于新增行时做骨架拷贝
  const [protoQuoteItem, setProtoQuoteItem] = useState<any>(null);

  // 1. 初始化加载全量数据表
  useEffect(() => {
    let active = true;
    const fetchData = async () => {
      setLoading(true);
      try {
        const res = await enterCreateQuoteApi();
        if (!active) return;

        // 将后端返回的 quote 结构体作为前端双向绑定的初始状态，同时回填系统默认值
        const initialQuote: Quote = {
          ...(res.data.quote || {}),
          valid_days: '30',
          creator_id: user?.id ? Number(user.id) : null,
          creator_name: user?.name || '未知经办人',
          quote_date: dayjs().format('YYYY-MM-DD'),
          customer_name: '',
          contact_name: '',
          contact_title: '',
        };
        setQuote(initialQuote);

        // 将后端返回的空 a_quote_item 结构体作为骨架，加 key 组装成单元素数组作为初始明细行
        const rawProtoItem = res.data.quote_item || {};
        setProtoQuoteItem(rawProtoItem);

        const defaultRow: FormQuoteItem = {
          ...rawProtoItem, // 保留后端传来的空结构体所有预设字段
          key: Math.random().toString(36).substring(2, 9),
          product_category_id: null,
          product_category_name: '',
          product_name_id: null,
          product_name: '',
          product_spec_id: null,
          product_spec: '',
          price_catalog_version: 'V1.0',
          order_batch_tier: '',
          catalog_base_price: 0,
          quote_float_rate: 0,
          quote_unit_price: 0,
          quantity: null,
          total_amount: 0,
          is_below_floor_price: false,
          high_threshold: 0,
          low_threshold: 0,
          big_batch_price: 0,
          middle_small_batch_price: 0,
          sample_small_order_price: 0,
        };
        setQuoteItems([defaultRow]); // 单元素数组初始化

        // 客户列表去重逻辑（过滤掉公司名为空的情况）
        const rawCustomers = res.data.customers || [];
        setAllCustomers(rawCustomers);

        const uniqueMap = new Map<string, CustomerRecord>();
        rawCustomers.forEach((item) => {
          if (item.company_name && item.company_name.trim()) {
            if (!uniqueMap.has(item.company_name)) {
              uniqueMap.set(item.company_name, item);
            }
          }
        });
        setUniqueCustomers(Array.from(uniqueMap.values()));

        // 映射产品规格宽表数据
        const specsWithExtra = (res.data.product_specs || []).map((spec: any) => ({
          ...spec,
          id: spec.id,
          product_name: spec.product_name || '',
          main_category: spec.main_category || '',
          spec: spec.spec || '',
          high_threshold: spec.high_threshold || 0,
          low_threshold: spec.low_threshold || 0,
          big_batch_price: Number(spec.big_batch_price || 0),
          middle_small_batch_price: Number(spec.middle_batch_price || 0),
          sample_small_order_price: Number(spec.small_batch_price || 0),
        }));
        setProductSpecsList(specsWithExtra);

        Feedback.handle(res, '基础数据加载成功');
      } catch (err: any) {
        if (!active) return;
        Feedback.handle(err, undefined, '加载基础数据失败');
      } finally {
        if (active) setLoading(false);
      }
    };

    fetchData();
    return () => {
      active = false;
    };
  }, [user]);

  // 2. 客户选择变动，双向绑定到 quote 状态中
  const handleCustomerChange = (val: string, isNew: boolean, record?: any) => {
    let customerId: number | null = null;
    let filtered: CustomerRecord[] = [];

    setIsNewCustomer(isNew);

    if (!isNew && val) {
      const actualRecord = record || uniqueCustomers.find((c) => c.company_name === val);
      if (actualRecord) {
        customerId = actualRecord.id;
      }
      filtered = allCustomers.filter(
        (c) =>
          c.company_name?.toLowerCase() === val.toLowerCase() &&
          c.contact_name &&
          c.contact_name.trim() !== ''
      );
    }

    setFilteredContacts(filtered);

    // 联动重置联系人和职位信息
    let soleContactName = '';
    let soleContactTitle = '';

    if (!isNew && filtered.length === 1) {
      soleContactName = filtered[0].contact_name || '';
      soleContactTitle = filtered[0].contact_title || '';
    }

    // 直接对整个 quote 结构体进行双向绑定更新
    setQuote((prev) => {
      if (!prev) return null;
      return {
        ...prev,
        customer_name: val,
        customer_id: customerId,
        contact_name: soleContactName,
        contact_title: soleContactTitle,
      };
    });
  };

  // 3. 联系人选择变动，双向绑定到 quote 状态中
  const handleContactChange = (val: string, isNew: boolean, record?: any) => {
    let title = '';
    if (!isNew) {
      if (record) {
        title = record.contact_title || '';
      } else {
        const found = filteredContacts.find((c) => c.contact_name === val);
        title = found?.contact_title || '';
      }
    }

    // 直接双向修改联系人与职位
    setQuote((prev) => {
      if (!prev) return null;
      return {
        ...prev,
        contact_name: val,
        contact_title: title,
      };
    });
  };

  // 4. 职位输入框手动输入，双向绑定到 quote 状态中
  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setQuote((prev) => {
      if (!prev) return null;
      return {
        ...prev,
        contact_title: val,
      };
    });
  };

  // 5. 报价明细行数据修改（双向状态修改 与 联动逻辑）
  const updateRow = (key: string, updatedFields: Partial<FormQuoteItem>) => {
    setQuoteItems((prev) =>
      prev.map((item) => {
        if (item.key !== key) return item;

        const newItem = { ...item, ...updatedFields };

        // 联动计算：根据批量档位，自动带出目录基准价并更新报价单价
        if (updatedFields.order_batch_tier !== undefined) {
          const tier = newItem.order_batch_tier;
          let base = 0;
          if (tier === '大批量') {
            base = newItem.big_batch_price || 0;
          } else if (tier === '中小批量') {
            base = newItem.middle_small_batch_price || 0;
          } else if (tier === '样品/小单') {
            base = newItem.sample_small_order_price || 0;
          }
          newItem.catalog_base_price = base;
          // 每次选择批量档位时，自动将该档位的目录基准价填入报价单价
          if (updatedFields.quote_unit_price === undefined) {
            newItem.quote_unit_price = base;
          }
        }

        // 联动计算：自动计算报价浮动比例 = ((报价单价 - 目录基准价) / 目录基准价) * 100
        const basePrice = newItem.catalog_base_price || 0;
        const unitPrice = newItem.quote_unit_price || 0;
        if (basePrice > 0 && unitPrice > 0) {
          newItem.quote_float_rate = Math.round(((unitPrice - basePrice) / basePrice) * 10000) / 100;
        } else {
          newItem.quote_float_rate = 0;
        }

        // 联动计算：合计数额 = 报价单价 * 数量
        const qty = newItem.quantity || 0;
        newItem.total_amount = Math.round(unitPrice * qty * 100) / 100;

        // 校验逻辑：数量范围合法性检查
        if (newItem.order_batch_tier && newItem.product_spec_id) {
          const qtyVal = newItem.quantity;
          if (qtyVal !== null) {
            let errStr = '';
            const low = newItem.low_threshold || 0;
            const high = newItem.high_threshold || 0;
            if (newItem.order_batch_tier === '大批量' && qtyVal < high) {
              errStr = `数量不规范，必须满大批量起订量 (≥ ${high})`;
            } else if (newItem.order_batch_tier === '中小批量' && (qtyVal < low || qtyVal >= high)) {
              errStr = `数量不在中小批量区间 [${low}, ${high - 1}] 内`;
            } else if (newItem.order_batch_tier === '样品/小单' && qtyVal >= low) {
              errStr = `数量超过了样品/小单上限 (< ${low})`;
            }
            newItem.quantityError = errStr;
          } else {
            newItem.quantityError = '';
          }
        } else {
          newItem.quantityError = '';
        }

        return newItem;
      })
    );
  };

  // 6. 添加新明细行
  const handleAddRow = () => {
    const newRow: FormQuoteItem = {
      ...protoQuoteItem, // 复制后端传来的原始空结构体骨架
      key: Math.random().toString(36).substring(2, 9),
      product_category_id: null,
      product_category_name: '',
      product_name_id: null,
      product_name: '',
      product_spec_id: null,
      product_spec: '',
      price_catalog_version: 'V1.0',
      order_batch_tier: '',
      catalog_base_price: 0,
      quote_float_rate: 0,
      quote_unit_price: 0,
      quantity: null,
      total_amount: 0,
      is_below_floor_price: false,
      high_threshold: 0,
      low_threshold: 0,
      big_batch_price: 0,
      middle_small_batch_price: 0,
      sample_small_order_price: 0,
    };
    setQuoteItems((prev) => [...prev, newRow]);
  };

  // 7. 删除明细行
  const handleDeleteRow = (key: string) => {
    if (quoteItems.length === 1) {
      Feedback.warning('必须保留至少一条明细记录');
      return;
    }
    setQuoteItems((prev) => prev.filter((it) => it.key !== key));
  };

  // 8. 提交整个报价单
  const handleFormSubmit = async () => {
    if (!quote) return;

    // 前端基础表单规则验证
    if (!quote.customer_name) {
      Feedback.error('请选择或填写客户名称');
      return;
    }

    if (quoteItems.length === 0) {
      Feedback.error('请添加至少一条报价明细行');
      return;
    }

    // 验证明细行的合规性
    for (let i = 0; i < quoteItems.length; i++) {
      const item = quoteItems[i];
      const lineNo = i + 1;
      if (!item.product_spec_id) {
        Feedback.error(`明细第 ${lineNo} 行：请选择产品名称`);
        return;
      }
      if (!item.order_batch_tier) {
        Feedback.error(`明细第 ${lineNo} 行：未选择批量档位`);
        return;
      }
      if (item.quote_unit_price === null || item.quote_unit_price === undefined || item.quote_unit_price <= 0) {
        Feedback.error(`明细第 ${lineNo} 行：请输入正确的报价单价`);
        return;
      }
      if (item.quantity === null || item.quantity <= 0) {
        Feedback.error(`明细第 ${lineNo} 行：请输入正确的产品数量`);
        return;
      }
      if (item.quantityError) {
        Feedback.error(`明细第 ${lineNo} 行：${item.quantityError}`);
        return;
      }
    }

    setSubmitting(true);

    try {
      // 直接打包双向绑定的数据发送给后端
      const payload: SubmitQuotePayload = {
        quote: {
          quote_code: quote.quote_code,
          customer_id: quote.customer_id !== undefined ? quote.customer_id : null,
          customer_name: quote.customer_name || '',
          contact_name: quote.contact_name || '',
          contact_title: quote.contact_title || '',
          valid_days: String(quote.valid_days || '30').replace('天', ''),
          creator_id: quote.creator_id !== undefined ? quote.creator_id : null,
          creator_name: quote.creator_name || '未知经办人',
          quote_date: dayjs(quote.quote_date).toISOString(),
        },
        quote_items: quoteItems.map((item) => ({
          product_category_id: item.product_category_id,
          product_category_name: item.product_category_name || '',
          product_name_id: item.product_name_id,
          product_name: item.product_name || '',
          product_spec_id: item.product_spec_id,
          product_spec: item.product_spec || '',
          price_catalog_version: item.price_catalog_version || 'V1.0',
          order_batch_tier: item.order_batch_tier || '',
          catalog_base_price: item.catalog_base_price || 0,
          quote_float_rate: (item.quote_float_rate || 0) / 100,
          quote_unit_price: item.quote_unit_price || 0,
          quantity: Number(item.quantity || 0),
          total_amount: item.total_amount || 0,
          is_below_floor_price: (item.quote_float_rate || 0) < 0,
        })),
        user: {
          id: user?.id ? Number(user.id) : 0,
          name: user?.name || '未知经办人',
        },
      };

      const res = await submitQuoteApi(payload);
      Feedback.handle(res, '报价单保存成功！');

      // 提交成功后重新拉取并重置数据
      const initData = await enterCreateQuoteApi();
      const nextQuote: Quote = {
        ...(initData.data.quote || {}),
        valid_days: '30',
        creator_id: user?.id ? Number(user.id) : null,
        creator_name: user?.name || '未知经办人',
        quote_date: dayjs().format('YYYY-MM-DD'),
        customer_name: '',
        contact_name: '',
        contact_title: '',
      };
      setQuote(nextQuote);

      // 重设单元素明细数组
      const rawProtoItem = initData.data.quote_item || {};
      setQuoteItems([{
        ...rawProtoItem,
        key: Math.random().toString(36).substring(2, 9),
        product_category_id: null,
        product_category_name: '',
        product_name_id: null,
        product_name: '',
        product_spec_id: null,
        product_spec: '',
        price_catalog_version: 'V1.0',
        order_batch_tier: '',
        catalog_base_price: 0,
        quote_float_rate: 0,
        quote_unit_price: 0,
        quantity: null,
        total_amount: 0,
        is_below_floor_price: false,
        high_threshold: 0,
        low_threshold: 0,
        big_batch_price: 0,
        middle_small_batch_price: 0,
        sample_small_order_price: 0,
      }]);

      setFilteredContacts([]);
      setIsNewCustomer(false);
    } catch (err: any) {
      Feedback.handle(err, undefined, '保存报价单时发生异常，请重试。');
    } finally {
      setSubmitting(false);
    }
  };

  // 9. 渲染表格列定义 (直接映射对应行的物理字段属性，通过 updateRow 触发状态修改)
  const columns = [
    {
      title: '产品名称 *',
      dataIndex: 'product_name',
      key: 'product_name',
      width: 200,
      render: (_text: string, record: FormQuoteItem) => {
        const selectVal = record.product_spec_id ? String(record.product_spec_id) : record.product_name;
        return (
          <DropdownMenu
            records={productSpecsList}
            displayField="product_name"
            valueField="id"
            value={selectVal}
            disableAddNew={true}
            onChange={(_val, _isNew, specRec) => {
              if (specRec) {
                updateRow(record.key, {
                  product_name: specRec.product_name,
                  product_spec_id: specRec.id,
                  product_name_id: null,
                  product_category_id: null,
                  product_category_name: specRec.main_category,
                  product_spec: specRec.spec || '',
                  high_threshold: specRec.high_threshold || 0,
                  low_threshold: specRec.low_threshold || 0,
                  big_batch_price: specRec.big_batch_price || 0,
                  middle_small_batch_price: specRec.middle_small_batch_price || 0,
                  sample_small_order_price: specRec.sample_small_order_price || 0,
                  order_batch_tier: '',
                  catalog_base_price: 0,
                  quote_unit_price: 0,
                });
              }
            }}
            placeholder="请选择产品"
          />
        );
      },
    },
    {
      title: '产品大类',
      dataIndex: 'product_category_name',
      key: 'product_category_name',
      width: 130,
      render: (text: string) => (
        <Input
          value={text}
          disabled={true}
          placeholder="自动带出"
        />
      ),
    },
    {
      title: '规格型号',
      dataIndex: 'product_spec',
      key: 'product_spec',
      width: 140,
      render: (text: string) => (
        <Input
          value={text}
          disabled={true}
          placeholder="自动带出"
        />
      ),
    },
    {
      title: '批量档位 *',
      dataIndex: 'order_batch_tier',
      key: 'order_batch_tier',
      width: 180,
      render: (text: string, record: FormQuoteItem) => {
        const hasSpec = record.product_spec_id !== null;
        return (
          <Select
            placeholder="选择批量档位"
            style={{ width: '100%' }}
            value={text || undefined}
            onChange={(val) => updateRow(record.key, { order_batch_tier: val })}
          >
            {hasSpec ? (
              <>
                <Option value="大批量">大批量 (≥{record.high_threshold})</Option>
                <Option value="中小批量">中小批量 ({record.low_threshold}~{record.high_threshold - 1})</Option>
                <Option value="样品/小单">样品/小单 (&lt;{record.low_threshold})</Option>
              </>
            ) : (
              <>
                <Option value="大批量">大批量</Option>
                <Option value="中小批量">中小批量</Option>
                <Option value="样品/小单">样品/小单</Option>
              </>
            )}
          </Select>
        );
      },
    },
    {
      title: '目录基准价',
      dataIndex: 'catalog_base_price',
      key: 'catalog_base_price',
      width: 110,
      render: (text: number) => <span>￥{text.toFixed(2)}</span>,
    },
    {
      title: '报价浮动比例',
      dataIndex: 'quote_float_rate',
      key: 'quote_float_rate',
      width: 120,
      render: (text: number) => {
        const val = text || 0;
        const formatted = val > 0 ? `+${val}%` : `${val}%`;
        const color = val < 0 ? '#cf1322' : val > 0 ? '#3f8600' : 'inherit';
        return <span style={{ color, fontWeight: 500 }}>{formatted}</span>;
      },
    },
    {
      title: '报价单价 *',
      dataIndex: 'quote_unit_price',
      key: 'quote_unit_price',
      width: 130,
      render: (text: number, record: FormQuoteItem) => (
        <InputNumber
          style={{ width: '100%' }}
          min={0}
          precision={2}
          prefix="￥"
          value={text || undefined}
          placeholder="填单价"
          onChange={(val) => updateRow(record.key, { quote_unit_price: val ?? 0 })}
        />
      ),
    },
    {
      title: '数量 *',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 140,
      render: (text: number, record: FormQuoteItem) => (
        <div>
          <InputNumber
            style={{ width: '100%' }}
            min={1}
            precision={0}
            value={text}
            placeholder="填数量"
            status={record.quantityError ? 'error' : ''}
            onChange={(val) => updateRow(record.key, { quantity: val })}
          />
          {record.quantityError && <span className="error-text">{record.quantityError}</span>}
        </div>
      ),
    },
    {
      title: '合计数额',
      dataIndex: 'total_amount',
      key: 'total_amount',
      width: 120,
      render: (text: number) => <strong style={{ color: '#1677ff' }}>￥{text.toFixed(2)}</strong>,
    },
    {
      title: '操作',
      key: 'action',
      width: 70,
      fixed: 'right' as const,
      render: (_text: any, record: FormQuoteItem) => (
        <Popconfirm
          title="确定删除此行明细吗？"
          okText="确定"
          cancelText="取消"
          onConfirm={() => handleDeleteRow(record.key)}
        >
          <Button type="text" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  if (loading) {
    return (
      <div className="create-quote-container" style={{ textAlign: 'center', paddingTop: '100px' }}>
        <Spin size="large" tip="正在加载新建报价单基础数据..." />
      </div>
    );
  }

  const hasCustomer = Boolean(quote?.customer_name && quote.customer_name.trim());
  const canEditContact = hasCustomer || isNewCustomer;

  return (
    <div className="create-quote-container">
      {/* 报价单基础信息卡片 */}
      <Card className="create-quote-card">
        <Title level={4} className="create-quote-title">
          报价单
        </Title>

        <Form layout="vertical">
          {/* 第一行 */}
          <Row gutter={24} className="form-row">
            <Col xs={24} sm={12} md={6}>
              <Form.Item label="报价单编号" className="custom-form-item">
                <Input
                  disabled
                  value={quote?.quote_code || ''}
                />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} md={6}>
              <Form.Item label="客户名称 *" className="custom-form-item">
                <DropdownMenu
                  records={uniqueCustomers}
                  displayField="company_name"
                  value={quote?.customer_name || ''}
                  onChange={handleCustomerChange}
                  placeholder="选择已有客户或新增填写"
                  inputPlaceholder="请输入客户名称"
                />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} md={6}>
              <Form.Item label="联系人" className="custom-form-item">
                <DropdownMenu
                  records={filteredContacts}
                  displayField="contact_name"
                  value={quote?.contact_name || ''}
                  onChange={handleContactChange}
                  placeholder={canEditContact ? "选择联系人或新增" : "请先选择或填写客户名称"}
                  inputPlaceholder={canEditContact ? "请输入联系人" : "请先选择或填写客户名称"}
                  disabled={!canEditContact}
                  forceNewMode={isNewCustomer ? true : undefined}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} md={6}>
              <Form.Item label="职位" className="custom-form-item">
                <Input
                  placeholder={canEditContact ? "请输入联系人职位" : "请先选择或填写客户名称"}
                  value={quote?.contact_title || ''}
                  onChange={handleTitleChange}
                  disabled={!canEditContact}
                />
              </Form.Item>
            </Col>
          </Row>

          {/* 第二行 */}
          <Row gutter={24} className="form-row">
            <Col xs={24} sm={12} md={8}>
              <Form.Item label="报价有效期 (天)" className="custom-form-item">
                <Input disabled value={quote?.valid_days || ''} />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} md={8}>
              <Form.Item label="市场部经办人" className="custom-form-item">
                <Input disabled value={quote?.creator_name || ''} />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} md={8}>
              <Form.Item label="报价日期" className="custom-form-item">
                <Input disabled value={quote?.quote_date ? dayjs(quote.quote_date).format('YYYY-MM-DD') : ''} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Card>

      {/* 报价明细列表卡片 */}
      <Card className="detail-table-card">
        <div className="detail-table-title">
          <Title level={4} style={{ margin: 0 }}>
            报价明细
          </Title>
          <Button 
            type="dashed" 
            onClick={handleAddRow} 
            icon={<PlusOutlined />}
            style={{ backgroundColor: '#e6f4ff', borderColor: '#91caff', color: '#0958d9' }}
          >
            添加明细行
          </Button>
        </div>

        <Table
          dataSource={quoteItems}
          columns={columns}
          pagination={false}
          scroll={{ x: 1200 }}
          bordered
          size="middle"
        />

        {/* 页面底部操作栏 */}
        <Row style={{ marginTop: '24px' }} justify="end">
          <Col>
            <Space size="middle">
              <span style={{ fontSize: '14px', color: '#555' }}>
                总明细项: <strong>{quoteItems.length}</strong> 项
              </span>
              <Button
                type="primary"
                size="large"
                loading={submitting}
                onClick={handleFormSubmit}
                icon={<SaveOutlined />}
                style={{ minWidth: '150px' }}
              >
                提交报价单
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>
    </div>
  );
};

export default CreateQuote;

