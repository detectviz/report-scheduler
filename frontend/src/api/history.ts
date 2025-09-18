import apiClient from './client';

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
  // apiClient 的回應攔截器會直接回傳 data，但 TypeScript 的靜態分析無法得知這一點。
  // 我們將其轉換為 unknown，然後再轉換為我們期望的型別，以通過型別檢查。
  const data = await apiClient.get(`/history?schedule_id=${scheduleId}`);
  return (data as unknown as HistoryLog[]).map(log => ({ ...log, key: log.id }));
};

/**
 * 觸發一次歷史紀錄的重寄
 * @param logId - 歷史紀錄的 UUID
 * @returns A promise that resolves to the success message.
 */
export const resendHistory = async (logId: string): Promise<{ message: string }> => {
  return apiClient.post(`/history/${logId}/resend`);
};
