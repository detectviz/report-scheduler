import apiClient from './client';
import { MOCK_ENABLED } from './mockConfig';
import { mockDataSources } from './mockData';

// 定義資料來源的資料結構 (與 models.DataSource 對應)
export interface DataSource {
    id: string;
    name: string;
    type: 'kibana' | 'grafana';
    url: string;
    api_url?: string;
    auth_type: 'basic_auth' | 'api_token' | 'none';
    credentials_ref?: string;
    version?: string;
    status: 'verified' | 'unverified' | 'error';
    created_at: string;
    updated_at: string;
}

// 對應後端的 models.AvailableElement
export interface AvailableElement {
    id: string;
    type: 'dashboard' | 'visualization' | 'saved_search';
    title: string;
}

// 獲取所有資料來源
export const getDataSources = (): Promise<DataSource[]> => {
    if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log('%c MOCKING API: getDataSources', 'color: #00b300');
        return new Promise(resolve => setTimeout(() => resolve([...mockDataSources]), 500));
    }
    return apiClient.get('/datasources');
};

// 根據 ID 獲取單一資料來源
export const getDataSourceById = (id: string): Promise<DataSource> => {
    if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log(`%c MOCKING API: getDataSourceById (id: ${id})`, 'color: #00b300');
        const ds = mockDataSources.find(d => d.id === id);
        return new Promise((resolve, reject) => setTimeout(() => ds ? resolve(ds) : reject(new Error('DataSource not found')), 300));
    }
    return apiClient.get(`/datasources/${id}`);
};

// 新增資料來源
export const createDataSource = (data: Partial<DataSource>): Promise<DataSource> => {
    if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log('%c MOCKING API: createDataSource', 'color: #00b300', data);
        const newDs: DataSource = {
            id: `ds-${Date.now()}`,
            name: data.name || 'New Mocked DS',
            type: data.type || 'kibana',
            url: data.url || '',
            auth_type: data.auth_type || 'none',
            status: 'unverified',
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
        };
        mockDataSources.push(newDs);
        return new Promise(resolve => setTimeout(() => resolve(newDs), 500));
    }
    return apiClient.post('/datasources', data);
};

// 更新資料來源
export const updateDataSource = (id: string, data: Partial<DataSource>): Promise<{ message: string }> => {
     if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log(`%c MOCKING API: updateDataSource (id: ${id})`, 'color: #00b300', data);
        const index = mockDataSources.findIndex(d => d.id === id);
        if (index !== -1) {
            mockDataSources[index] = { ...mockDataSources[index], ...data, updated_at: new Date().toISOString() };
        }
        return new Promise(resolve => setTimeout(() => resolve({ message: '資料來源已成功更新' }), 500));
    }
    return apiClient.put(`/datasources/${id}`, data);
};

// 刪除資料來源
export const deleteDataSource = (id: string): Promise<{ message: string }> => {
    if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log(`%c MOCKING API: deleteDataSource (id: ${id})`, 'color: #00b300');
        // In a real mock setup, you might filter the array. For now, we just resolve.
        return new Promise(resolve => setTimeout(() => resolve({ message: '資料來源已成功刪除' }), 500));
    }
    return apiClient.delete(`/datasources/${id}`);
};

// 驗證資料來源連線
export const validateDataSource = (id: string): Promise<{ status: string; message: string }> => {
    if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log(`%c MOCKING API: validateDataSource (id: ${id})`, 'color: #00b300');
        const ds = mockDataSources.find(d => d.id === id);
        if (ds) ds.status = 'verified';
        return new Promise(resolve => setTimeout(() => resolve({ status: 'verified', message: '連線成功' }), 1000));
    }
    return apiClient.post(`/datasources/${id}/validate`);
};

// 獲取指定資料來源下的可用元素
export const getDataSourceElements = (dataSourceId: string): Promise<AvailableElement[]> => {
    if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log(`%c MOCKING API: getDataSourceElements (id: ${dataSourceId})`, 'color: #00b300');
        const mockElements: AvailableElement[] = [
            { id: 'element-1', type: 'dashboard', title: 'Sample Dashboard 1' },
            { id: 'element-2', type: 'visualization', title: 'Sales Chart' },
            { id: 'element-3', type: 'saved_search', title: 'Error Logs' },
        ];
        return new Promise(resolve => setTimeout(() => resolve(mockElements), 800));
    }
    return apiClient.get(`/datasources/${dataSourceId}/elements`);
};
