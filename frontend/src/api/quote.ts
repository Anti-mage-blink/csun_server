import request from './request';

export interface CustomerRecord {
  id: number;
  company_name: string | null;
  contact_name: string | null;
  contact_title: string | null;
}

export interface PrepareCreateQuoteData {
  quote_code: string;
  customers: CustomerRecord[];
  product_categories: any[];
  product_names: any[];
  product_specs: any[];
  price_catalogs: any[];
}

export interface PrepareCreateQuoteResponse {
  message: string;
  data: PrepareCreateQuoteData;
}

/**
 * 进入新建报价单页面，获取报价单编号、客户列表及相关全量基础数据
 */
export const enterCreateQuoteApi = async (): Promise<PrepareCreateQuoteResponse> => {
  try {
    const res = await request.get<PrepareCreateQuoteResponse>('/quote/create');
    return res.data;
  } catch (error: any) {
    console.warn('连接后端 /quote/create 失败，可能后端服务未启动。正在自动降级至前端 Mock 运行环境。');
    
    // 如果后端未启动，降级返回 Mock 数据
    return {
      message: '获取新建报价单数据成功 (Mock 降级)',
      data: {
        quote_code: `BJ-${new Date().toISOString().slice(0, 10).replace(/-/g, '')}-001`,
        customers: [
          { id: 1, company_name: '比亚迪股份有限公司', contact_name: '王总', contact_title: '采购总监' },
          { id: 2, company_name: '比亚迪股份有限公司', contact_name: '张经理', contact_title: '采购经理' },
          { id: 3, company_name: '宁德时代新能源科技', contact_name: '曾董', contact_title: '董事长' },
          { id: 4, company_name: '宁德时代新能源科技', contact_name: '李专员', contact_title: '采购助理' },
          { id: 5, company_name: '小米汽车', contact_name: '雷总', contact_title: 'CEO' }
        ],
        product_categories: [],
        product_names: [],
        product_specs: [],
        price_catalogs: []
      }
    };
  }
};
