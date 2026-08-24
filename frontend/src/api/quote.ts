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
  pay_way?: string | null;
  credit_period?: string | null;
  remarks?: string | null;
  attachment_path_array?: string | string[] | null;
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
 * 进入发起报价单页面，获取报价单编号、客户列表及相关全量基础数据
 */
export const enterCreateQuoteApi = async (): Promise<PrepareCreateQuoteResponse> => {
  try {
    const res = await request.get<PrepareCreateQuoteResponse>('/quote/create');
    return res.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '获取发起报价单数据失败');
    }
    throw new Error(error.message || '连接服务器失败，获取发起报价单数据失败');
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
    pay_way: string;
    credit_period: string;
    remarks: string;
    attachment_path_array?: string[] | null;
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
    is_below_floor_price: boolean;
  }>;
  user: {
    id: number;
    name: string;
  };
}

/**
 * 提交发起报价单
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
  creator_id: number | null;
  creator_name: string | null;
  approver_id?: number | null;
  approver_name?: string | null;
  present_approver_id?: number | null;
  present_approver_name?: string | null;
  present_node_id?: number | null;
  present_node_name?: string | null;
  present_status?: string | null;
  created_at?: string | null;
  updated_at?: string | null;
}

export interface QuoteProcessNode {
  id: number;
  process_id: number | null;
  seq_num?: number | null;
  name: string | null;
  approver_id: number | null;
  approver_name: string | null;
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
  is_below_floor_price?: boolean | null;
}

export interface MyApproveQueryData {
  total: number;
  quote_processes: QuoteProcess[];
  quote_process_nodes: QuoteProcessNode[];
  quotes: Quote[];
  quote_items: QuoteItem[];
}

export interface MyApproveQueryResponse {
  message: string;
  data: MyApproveQueryData;
}

/**
 * 查询待审批数据
 */
export const myApproveQueryApi = async (userId: number): Promise<MyApproveQueryResponse> => {
  try {
    const res = await request.get<MyApproveQueryResponse>('/approve/my_approve_query', {
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
 * 查询我的发起数据
 */
export const myApplyQueryApi = async (userId: number): Promise<MyApplyQueryResponse> => {
  try {
    const res = await request.get<MyApplyQueryResponse>('/apply/my_apply_query', {
      params: { user_id: userId }
    });
    return res.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '查询我的发起数据失败');
    }
    throw new Error(error.message || '连接服务器失败，查询我的发起数据失败');
  }
};

export interface ApproveUser {
  id: number;
  name: string;
  role: string;
}

export interface ApproveHandlePayload {
  action: 'approve' | 'reject';
  node_id: number;
  process_id: number;
  comment: string;
  user?: ApproveUser;
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

export interface WithdrawQuotePayload {
  process_id: number;
  user: {
    id: number;
    name: string;
  };
}

export interface WithdrawQuoteResponse {
  message: string;
}

/**
 * 撤回报价单接口
 */
export const withdrawQuoteApi = async (payload: WithdrawQuotePayload): Promise<WithdrawQuoteResponse> => {
  try {
    const res = await request.post<WithdrawQuoteResponse>('/quote/withdraw_quote', payload);
    return res.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw new Error(error.response.data.message || '撤回报价单失败');
    }
    throw new Error(error.message || '连接服务器失败，撤回报价单失败');
  }
};

