import apiClient from './client';
import { MOCK_ENABLED } from './mockConfig';
import { mockReportDefinitions } from './mockData';

// 對應後端的 models.ReportElement
export interface ReportElement {
    id: string;
    type: 'dashboard' | 'visualization' | 'saved_search';
    title: string;
    order: number;
}

// 對應後端的 models.ReportDefinition
export interface ReportDefinition {
    id: string;
    name: string;
    description?: string;
    datasource_id: string;
    time_range: string;
    elements: ReportElement[];
    created_at: string;
    updated_at: string;
}

// 獲取所有報表定義
export const getReportDefinitions = (): Promise<ReportDefinition[]> => {
    if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log('%c MOCKING API: getReportDefinitions', 'color: #00b300');
        return new Promise(resolve => setTimeout(() => resolve([...mockReportDefinitions]), 500));
    }
    return apiClient.get('/reports');
};

// 根據 ID 獲取單一報表定義
export const getReportDefinitionById = (id: string): Promise<ReportDefinition> => {
    if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log(`%c MOCKING API: getReportDefinitionById (id: ${id})`, 'color: #00b300');
        const report = mockReportDefinitions.find(r => r.id === id);
        return new Promise((resolve, reject) => setTimeout(() => report ? resolve(report) : reject(new Error('ReportDefinition not found')), 300));
    }
    return apiClient.get(`/reports/${id}`);
};

// 新增報表定義
export const createReportDefinition = (data: Partial<ReportDefinition>): Promise<ReportDefinition> => {
    if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log('%c MOCKING API: createReportDefinition', 'color: #00b300', data);
        const newReport: ReportDefinition = {
            id: `report-${Date.now()}`,
            name: data.name || 'New Mocked Report',
            datasource_id: data.datasource_id || 'ds-1',
            time_range: data.time_range || 'now-1h',
            elements: data.elements || [],
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
        };
        mockReportDefinitions.push(newReport);
        return new Promise(resolve => setTimeout(() => resolve(newReport), 500));
    }
    return apiClient.post('/reports', data);
};

// 更新報表定義
export const updateReportDefinition = (id: string, data: Partial<ReportDefinition>): Promise<{ message: string }> => {
    if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log(`%c MOCKING API: updateReportDefinition (id: ${id})`, 'color: #00b300', data);
        return new Promise(resolve => setTimeout(() => resolve({ message: '報表定義已成功更新' }), 500));
    }
    return apiClient.put(`/reports/${id}`, data);
};

// 刪除報表定義
export const deleteReportDefinition = (id: string): Promise<{ message: string }> => {
    if (import.meta.env.DEV && MOCK_ENABLED) {
        console.log(`%c MOCKING API: deleteReportDefinition (id: ${id})`, 'color: #00b300');
        return new Promise(resolve => setTimeout(() => resolve({ message: '報表定義已成功刪除' }), 500));
    }
    return apiClient.delete(`/reports/${id}`);
};
