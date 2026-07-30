import request from './request';

export interface DataTableQueryResult<T = any> {
  message?: string;
  data: T[];
  relations?: Record<string, any[]>;
}

/**
 * 查：获取主表全量记录以及所有关联表的全量记录列表
 * @param table 主表标识字符串，格式如 "general.customer" 或 "customer"
 * @param relationTables 关联表标识字符串数组，格式如 ["quote_manage.product_spec"]
 */
export const fetchDataTableData = async <T = any>(
  table: string,
  relationTables?: string[]
): Promise<DataTableQueryResult<T>> => {
  try {
    const params: Record<string, any> = { table };
    if (relationTables && relationTables.length > 0) {
      params.relation_tables = relationTables.join(',');
    }

    const res = await request.get<DataTableQueryResult<T>>('/data_table', { params });

    // 兼容多种返回结构
    if (res.data && Array.isArray(res.data.data)) {
      return {
        message: res.data.message,
        data: res.data.data,
        relations: res.data.relations || {},
      };
    }

    if (Array.isArray(res.data)) {
      return {
        data: res.data,
        relations: {},
      };
    }

    return {
      data: (res.data as any)?.data || [],
      relations: (res.data as any)?.relations || {},
    };
  } catch (error: any) {
    const msg = error.response?.data?.message || error.message || '查询数据失败';
    throw new Error(msg);
  }
};

/**
 * 增：写入一条新记录
 * @param table 主表标识字符串，如 "general.customer"
 * @param record 1:1 结构体数据
 */
export const createDataTableRecord = async (
  table: string,
  record: Record<string, any>
): Promise<any> => {
  try {
    const res = await request.post('/data_table', {
      table,
      record,
    });
    return res.data;
  } catch (error: any) {
    const msg = error.response?.data?.message || error.message || '新增记录失败';
    throw new Error(msg);
  }
};

/**
 * 改：根据记录 id 修改记录
 * @param table 主表标识字符串，如 "general.customer"
 * @param id 目标记录主键 ID
 * @param record 1:1 结构体数据
 */
export const updateDataTableRecord = async (
  table: string,
  id: number | string,
  record: Record<string, any>
): Promise<any> => {
  try {
    const res = await request.put('/data_table', {
      table,
      id,
      record,
    });
    return res.data;
  } catch (error: any) {
    const msg = error.response?.data?.message || error.message || '修改记录失败';
    throw new Error(msg);
  }
};

/**
 * 删：软删除一条记录
 * @param table 主表标识字符串，如 "general.customer"
 * @param id 目标记录主键 ID
 */
export const deleteDataTableRecord = async (
  table: string,
  id: number | string
): Promise<any> => {
  try {
    const res = await request.delete('/data_table', {
      data: {
        table,
        id,
      },
    });
    return res.data;
  } catch (error: any) {
    const msg = error.response?.data?.message || error.message || '删除记录失败';
    throw new Error(msg);
  }
};
