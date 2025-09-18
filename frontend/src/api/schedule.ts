import apiClient from './client';

// 對應後端的 models.Recipients
export interface Recipients {
  to: string[];
  cc?: string[];
  bcc?: string[];
}

// 對應後端的 models.Schedule
export interface Schedule {
  id: string;
  name: string;
  cron_spec: string;
  timezone: string;
  recipients: Recipients;
  email_subject: string;
  email_body: string;
  report_ids: string[];
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
}

// 獲取所有排程
export const getSchedules = (): Promise<Schedule[]> => {
  return apiClient.get('/schedules');
};

// 根據 ID 獲取單一排程
export const getScheduleById = (id: string): Promise<Schedule> => {
  return apiClient.get(`/schedules/${id}`);
};

// 新增排程
export const createSchedule = (data: Partial<Schedule>): Promise<Schedule> => {
  return apiClient.post('/schedules', data);
};

// 更新排程
export const updateSchedule = (id: string, data: Partial<Schedule>): Promise<{ message: string }> => {
  return apiClient.put(`/schedules/${id}`, data);
};

// 刪除排程
export const deleteSchedule = (id: string): Promise<{ message: string }> => {
  return apiClient.delete(`/schedules/${id}`);
};

// 手動觸發排程
export const triggerSchedule = (id: string): Promise<{ message: string; task_id: string }> => {
  return apiClient.post(`/schedules/${id}/trigger`);
};
