import request from './request';

export interface CustomerRecord {
  id: number;
  company_name: string | null;
  contact_name: string | null;
  contact_title: string | null;
}

export interface Quote {
  id?: number;
  quote_code: string;
  customer_id?: number | null;
  customer_name?: string | null;
  contact_name?: string | null;
  contact_title?: string | null;
  valid_days?: string | null;
  creator_id?: number | null;
  creator_name?: string | null;
  quote_date?: string | null;
}

export interface PrepareCreateQuoteData {
  quote: Quote;
  quote_item?: any;
  customers: CustomerRecord[];
  product_specs: any[];
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
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '获取新建报价单数据失败');
    }
    throw new Error(error.message || '连接服务器失败，获取新建报价单数据失败');
  }
};

export interface SubmitQuotePayload {
  quote: {
    quote_code: string;
    customer_id: number | null;
    customer_name: string;
    contact_name: string;
    contact_title: string;
    valid_days: string;
    creator_id: number | null;
    creator_name: string;
    quote_date: string;
  };
  quote_items: Array<{
    product_category_id: number | null;
    product_category_name: string;
    product_name_id: number | null;
    product_name: string;
    product_spec_id: number | null;
    product_spec: string;
    price_catalog_version: string;
    order_batch_tier: string;
    catalog_base_price: number;
    quote_float_rate: number;
    quote_unit_price: number;
    quantity: number;
    total_amount: number;
    is_below_price_floor: boolean;
  }>;
  user: {
    id: number;
    name: string;
  };
}

/**
 * 提交新建报价单
 */
export const submitQuoteApi = async (payload: SubmitQuotePayload): Promise<{ message: string }> => {
  try {
    const res = await request.post<{ message: string }>('/quote/submit', payload);
    return res.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '提交报价单失败');
    }
    throw new Error(error.message || '连接服务器失败，提交报价单失败');
  }
};

export interface QuoteProcess {
  id: number;
  quote_id: number | null;
  create_employee_id: number | null;
  create_employee_name: string | null;
  present_approver_id?: number | null;
  present_approver_name?: string | null;
  present_node_id?: number | null;
}

export interface QuoteProcessNode {
  id: number;
  process_id: number | null;
  seq_num?: number | null;
  name: string | null;
  approve_employee_id: number | null;
  approve_employee_name: string | null;
  status: string | null;
  approve_comment: string | null;
  created_at: string | null;
  approve_at: string | null;
}

export interface QuoteItem {
  id: number;
  quote_id?: number | null;
  product_category_id?: number | null;
  product_category_name?: string | null;
  product_name_id?: number | null;
  product_name?: string | null;
  product_spec_id?: number | null;
  product_spec?: string | null;
  price_catalog_version?: string | null;
  order_batch_tier?: string | null;
  catalog_base_price?: number | null;
  quote_float_rate?: number | null;
  quote_unit_price?: number | null;
  quantity?: number | null;
  total_amount?: number | null;
  is_below_price_floor?: boolean | null;
}

export interface QueryNeedApproveData {
  total: number;
  quote_processes: QuoteProcess[];
  quote_process_nodes: QuoteProcessNode[];
  quotes: Quote[];
  quote_items: QuoteItem[];
}

export interface QueryNeedApproveResponse {
  message: string;
  data: QueryNeedApproveData;
}

/**
 * 查询待审批数据
 */
export const queryNeedApproveApi = async (userId: number): Promise<QueryNeedApproveResponse> => {
  try {
    const res = await request.get<QueryNeedApproveResponse>('/approve/query_need_approve', {
      params: { user_id: userId }
    });
    return res.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '查询待审批数据失败');
    }
    throw new Error(error.message || '连接服务器失败，查询待审批数据失败');
  }
};

export interface FilingLookResponse {
  message: string;
  data: {
    quote_processes: QuoteProcess[];
    quote_process_nodes: QuoteProcessNode[];
    quotes: Quote[];
    quote_items: QuoteItem[];
  };
}

/**
 * 备案查看，获取全量报价流程相关数据
 */
export const filingLookApi = async (): Promise<FilingLookResponse> => {
  try {
    const res = await request.get<FilingLookResponse>('/filing/filing_look');
    return res.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '备案查看获取数据失败');
    }
    throw new Error(error.message || '连接服务器失败，备案查看获取数据失败');
  }
};

export interface MyApplyQueryResponse {
  message: string;
  data: {
    quote_processes: QuoteProcess[];
    quote_process_nodes: QuoteProcessNode[];
    quotes: Quote[];
    quote_items: QuoteItem[];
  };
}

/**
 * 查询我的申请数据
 */
export const myApplyQueryApi = async (userId: number): Promise<MyApplyQueryResponse> => {
  try {
    const res = await request.get<MyApplyQueryResponse>('/apply/my_apply_query', {
      params: { user_id: userId }
    });
    return res.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '查询我的申请数据失败');
    }
    throw new Error(error.message || '连接服务器失败，查询我的申请数据失败');
  }
};

export interface ApproveHandlePayload {
  action: 'approve' | 'reject';
  node_id: number;
  process_id: number;
  comment: string;
}

export interface ApproveHandleResponse {
  message: string;
}

/**
 * 审批处理接口：同意通过（approve）或 拒绝退回（reject）
 */
export const approveHandleApi = async (payload: ApproveHandlePayload): Promise<ApproveHandleResponse> => {
  try {
    const res = await request.post<ApproveHandleResponse>('/approve/approve_handle', payload);
    return res.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '审批操作处理失败');
    }
    throw new Error(error.message || '连接服务器失败，审批操作处理失败');
  }
};
