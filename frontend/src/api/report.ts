import apiClient from './client';

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
    return apiClient.post('/reports', data);
};

// 更新報表定義
export const updateReportDefinition = (id: string, data: Partial<ReportDefinition>): Promise<{ message: string }> => {
    return apiClient.put(`/reports/${id}`, data);
};

// 刪除報表定義
export const deleteReportDefinition = (id: string): Promise<{ message: string }> => {
    return apiClient.delete(`/reports/${id}`);
};
