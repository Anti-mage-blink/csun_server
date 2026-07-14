import React, { useState, useEffect } from 'react';
import { Form, Row, Col, Input, Typography, Card, Spin, message, Button } from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { useAuth } from '@/context/AuthContext';
import DropdownMenu from '@/components/dropdownMenu';
import { enterCreateQuoteApi, type CustomerRecord } from '@/api/quote';
import './index.css';

const { Title } = Typography;

/**
 * 新建报价单模块
 */
const CreateQuote: React.FC = () => {
  const [form] = Form.useForm();
  const { user } = useAuth();
  const [loading, setLoading] = useState(true);

  // 数据源状态
  const [allCustomers, setAllCustomers] = useState<CustomerRecord[]>([]); // 后端返回的所有客户/联系人记录
  const [uniqueCustomers, setUniqueCustomers] = useState<CustomerRecord[]>([]); // 去重后的客户记录 (以 company_name 去重)
  const [filteredContacts, setFilteredContacts] = useState<CustomerRecord[]>([]); // 过滤后的联系人记录

  // 当前选中的表单字段状态
  const [selectedCustomer, setSelectedCustomer] = useState('');
  const [selectedContact, setSelectedContact] = useState('');
  const [contactTitle, setContactTitle] = useState('');

  // 初始化加载数据
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const res = await enterCreateQuoteApi();
        
        // 1. 设置报价单编号及默认值
        form.setFieldsValue({
          quote_code: res.data.quote_code,
          valid_days: '30天',
          handler_name: user?.username || '未知经办人',
          quote_date: dayjs().format('YYYY-MM-DD'),
        });

        // 2. 保存原始客户列表
        const rawCustomers = res.data.customers || [];
        setAllCustomers(rawCustomers);

        // 3. 对客户进行公司名称去重，以便客户名称下拉菜单展示
        const uniqueMap = new Map<string, CustomerRecord>();
        rawCustomers.forEach((item) => {
          if (item.company_name) {
            // 如果 map 中没有或者当前项包含更多有效信息，则存入
            if (!uniqueMap.has(item.company_name)) {
              uniqueMap.set(item.company_name, item);
            }
          }
        });
        setUniqueCustomers(Array.from(uniqueMap.values()));
        
        message.success('报价单基础数据加载成功');
      } catch (err: any) {
        message.error('加载报价单初始化数据失败');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [form, user]);

  // 当客户名称变动时的回调
  const handleCustomerChange = (val: string, isNew: boolean) => {
    setSelectedCustomer(val);
    form.setFieldsValue({ customer_name: val });

    // 联动重置联系人和职位
    setSelectedContact('');
    setContactTitle('');
    form.setFieldsValue({
      contact_name: '',
      contact_title: '',
    });

    if (isNew || !val) {
      // 如果是“新增填写”或清空，清空联系人过滤列表
      setFilteredContacts([]);
    } else {
      // 过滤出当前公司名下的所有联系人列表
      const filtered = allCustomers.filter(
        (c) => c.company_name?.toLowerCase() === val.toLowerCase()
      );
      setFilteredContacts(filtered);

      // 如果该公司只有唯一的一个联系人，则可以默认自动带出
      if (filtered.length === 1) {
        const soleContact = filtered[0];
        const soleName = soleContact.contact_name || '';
        const soleTitle = soleContact.contact_title || '';
        setSelectedContact(soleName);
        setContactTitle(soleTitle);
        form.setFieldsValue({
          contact_name: soleName,
          contact_title: soleTitle,
        });
      }
    }
  };

  // 当联系人变动时的回调
  const handleContactChange = (val: string, isNew: boolean, record?: any) => {
    setSelectedContact(val);
    form.setFieldsValue({ contact_name: val });

    if (isNew) {
      // 新增联系人时，清空职位，允许用户自行在职位输入框输入
      setContactTitle('');
      form.setFieldsValue({ contact_title: '' });
    } else if (record) {
      // 选择已有联系人时，自动带出职位
      const title = record.contact_title || '';
      setContactTitle(title);
      form.setFieldsValue({ contact_title: title });
    } else {
      // 根据联系人名称从已过滤的联系人列表中寻找职位
      const found = filteredContacts.find((c) => c.contact_name === val);
      const title = found?.contact_title || '';
      setContactTitle(title);
      form.setFieldsValue({ contact_title: title });
    }
  };

  // 职位输入框变动时的回调
  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setContactTitle(val);
    form.setFieldsValue({ contact_title: val });
  };

  // 提交新建报价单
  const onFinish = (values: any) => {
    console.log('提交的报价单表单数据：', values);
    message.success('报价单提交成功！(联调阶段将对接保存接口)');
  };

  if (loading) {
    return (
      <div className="create-quote-container" style={{ textAlign: 'center', paddingTop: '100px' }}>
        <Spin size="large" tip="正在加载新建报价单初始化数据..." />
      </div>
    );
  }

  return (
    <div className="create-quote-container">
      <Card className="create-quote-card">
        <Title level={4} className="create-quote-title">
          报价单
        </Title>

        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            valid_days: '30天',
            quote_date: dayjs().format('YYYY-MM-DD'),
          }}
        >
          {/* 第一行 */}
          <Row gutter={24} className="form-row">
            <Col xs={24} sm={12} md={6}>
              <Form.Item
                name="quote_code"
                label="报价单编号"
                className="custom-form-item"
              >
                <Input disabled />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} md={6}>
              <Form.Item
                name="customer_name"
                label="客户名称"
                className="custom-form-item"
                rules={[{ required: true, message: '请选择或填写客户名称' }]}
              >
                <DropdownMenu
                  records={uniqueCustomers}
                  displayField="company_name"
                  value={selectedCustomer}
                  onChange={handleCustomerChange}
                  placeholder="选择客户或新增"
                />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} md={6}>
              <Form.Item
                name="contact_name"
                label="联系人"
                className="custom-form-item"
                rules={[{ required: true, message: '请选择或填写联系人' }]}
              >
                <DropdownMenu
                  records={filteredContacts}
                  displayField="contact_name"
                  value={selectedContact}
                  onChange={handleContactChange}
                  placeholder={selectedCustomer ? "选择联系人或新增" : "请先选择客户"}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} md={6}>
              <Form.Item
                name="contact_title"
                label="职位"
                className="custom-form-item"
              >
                <Input
                  placeholder="联系人职位 (自动带出/可输入)"
                  value={contactTitle}
                  onChange={handleTitleChange}
                />
              </Form.Item>
            </Col>
          </Row>

          {/* 第二行 */}
          <Row gutter={24} className="form-row">
            <Col xs={24} sm={12} md={8}>
              <Form.Item
                name="valid_days"
                label="报价有效期"
                className="custom-form-item"
              >
                <Input disabled />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} md={8}>
              <Form.Item
                name="handler_name"
                label="市场部经办人"
                className="custom-form-item"
              >
                <Input disabled />
              </Form.Item>
            </Col>

            <Col xs={24} sm={12} md={8}>
              <Form.Item
                name="quote_date"
                label="报价日期"
                className="custom-form-item"
              >
                <Input disabled />
              </Form.Item>
            </Col>
          </Row>

          {/* 底部操作按钮 */}
          <Row style={{ marginTop: '24px' }} justify="end">
            <Col>
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />}>
                提交报价单
              </Button>
            </Col>
          </Row>
        </Form>
      </Card>
    </div>
  );
};

export default CreateQuote;
