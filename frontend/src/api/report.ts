import apiClient from './client';
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
    return apiClient.get('/reports');
};

// 根據 ID 獲取單一報表定義
export const getReportDefinitionById = (id: string): Promise<ReportDefinition> => {
    return apiClient.get(`/reports/${id}`);
};

// 新增報表定義
export const createReportDefinition = (data: Partial<ReportDefinition>): Promise<ReportDefinition> => {
    if (import.meta.env.VITE_MOCK_ENABLED === 'true') {
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
    if (import.meta.env.VITE_MOCK_ENABLED === 'true') {
        return new Promise(resolve => setTimeout(() => resolve({ message: '報表定義已成功更新' }), 500));
    }
    return apiClient.put(`/reports/${id}`, data);
};

// 刪除報表定義
export const deleteReportDefinition = (id: string): Promise<{ message: string }> => {
    if (import.meta.env.VITE_MOCK_ENABLED === 'true') {
        return new Promise(resolve => setTimeout(() => resolve({ message: '報表定義已成功刪除' }), 500));
    }
    return apiClient.delete(`/reports/${id}`);
};

// 同步產生並預覽報表
export const generateReportPreview = (report: ReportDefinition): Promise<{ preview_url: string }> => {
    // 條件式模擬：當 MOCK_ENABLED 為 true，但資料來源不是 ds-4 (公開 Kibana) 時，才使用模擬
    if (import.meta.env.VITE_MOCK_ENABLED === 'true' && report.datasource_id !== 'ds-4') {
        const mockPdfUrl = 'https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf';
        console.log(`使用模擬 API (資料來源: ${report.datasource_id})`);
        return new Promise(resolve => setTimeout(() => resolve({ preview_url: mockPdfUrl }), 2000));
    }
    // 對於 ds-4 或 MOCK_ENABLED 為 false 的情況，呼叫真實 API
    console.log(`呼叫真實 API (資料來源: ${report.datasource_id})`);
    return apiClient.post(`/reports/${report.id}/generate`);
};
