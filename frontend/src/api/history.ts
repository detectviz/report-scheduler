import apiClient from './client';
import { mockHistoryLogs } from './mockData';

// 根據 specs.md 和後端 models 定義 HistoryLog 的 TypeScript 型別
export interface HistoryLog {
  id: string;
  schedule_id: string;
  schedule_name: string;
  trigger_time: string; // ISO 8601 date string
  execution_duration_ms: number;
  status: 'success' | 'error' | 'retrying';
  error_message?: string;
  recipients: string; // JSON string
  report_url?: string;
  key?: string; // antd table 需要的 key
}

/**
 * 根據排程 ID 獲取歷史紀錄
 * @param scheduleId - 排程的 UUID
 * @returns A promise that resolves to an array of history logs.
 */
export const getHistoryByScheduleId = async (scheduleId: string): Promise<HistoryLog[]> => {
  if (import.meta.env.VITE_MOCK_ENABLED === 'true') {
    console.log(`%c MOCKING API: getHistoryByScheduleId (scheduleId: ${scheduleId})`, 'color: #00b300');
    const logs = mockHistoryLogs.filter(log => log.schedule_id === scheduleId);
    return new Promise(resolve => setTimeout(() => resolve(logs.map(log => ({...log, key: log.id}))), 500));
  }
  const data = await apiClient.get(`/history?schedule_id=${scheduleId}`);
  // The response interceptor returns data directly. Cast to unknown first to satisfy TS.
  return (data as unknown as HistoryLog[]).map(log => ({ ...log, key: log.id }));
};

/**
 * 觸發一次歷史紀錄的重寄
 * @param logId - 歷史紀錄的 UUID
 * @returns A promise that resolves to the success message.
 */
export const resendHistory = async (logId: string): Promise<{ message: string }> => {
  if (import.meta.env.VITE_MOCK_ENABLED === 'true') {
    console.log(`%c MOCKING API: resendHistory (logId: ${logId})`, 'color: #00b300');
    return new Promise(resolve => setTimeout(() => resolve({ message: '已成功將重寄任務加入佇列' }), 500));
  }
  return apiClient.post(`/history/${logId}/resend`);
};
