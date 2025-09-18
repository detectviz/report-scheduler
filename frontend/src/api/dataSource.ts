import apiClient from './client';

// 定義資料來源的資料結構 (與 models.DataSource 對應)
export interface DataSource {
    id: string;
    name: string;
    type: 'kibana' | 'grafana';
    url: string;
    api_url?: string;
    auth_type: 'basic_auth' | 'api_token';
    credentials_ref?: string;
    version?: string;
    status: 'verified' | 'unverified' | 'error';
    created_at: string;
    updated_at: string;
}

// 獲取所有資料來源
export const getDataSources = (): Promise<DataSource[]> => {
    return apiClient.get('/datasources');
};

// 根據 ID 獲取單一資料來源
export const getDataSourceById = (id: string): Promise<DataSource> => {
    return apiClient.get(`/datasources/${id}`);
};

// 新增資料來源
export const createDataSource = (data: Partial<DataSource>): Promise<DataSource> => {
    return apiClient.post('/datasources', data);
};

// 更新資料來源
export const updateDataSource = (id: string, data: Partial<DataSource>): Promise<{ message: string }> => {
    return apiClient.put(`/datasources/${id}`, data);
};

// 刪除資料來源
export const deleteDataSource = (id: string): Promise<{ message: string }> => {
    return apiClient.delete(`/datasources/${id}`);
};

// 驗證資料來源連線
export const validateDataSource = (id: string): Promise<{ status: string; message: string }> => {
    return apiClient.post(`/datasources/${id}/validate`);
};

// 對應後端的 models.AvailableElement
export interface AvailableElement {
    id: string;
    type: 'dashboard' | 'visualization' | 'saved_search';
    title: string;
}

// 獲取指定資料來源下的可用元素
export const getDataSourceElements = (dataSourceId: string): Promise<AvailableElement[]> => {
    return apiClient.get(`/datasources/${dataSourceId}/elements`);
};
