import apiClient from './client';
import { mockSchedules } from './mockData';

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
  email_subject?: string;
  email_body?: string;
  report_ids: string[];
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
}

// 獲取所有排程
export const getSchedules = (): Promise<Schedule[]> => {
  if (import.meta.env.VITE_MOCK_ENABLED === 'true') {
    console.log('%c MOCKING API: getSchedules', 'color: #00b300');
    return new Promise(resolve => setTimeout(() => resolve([...mockSchedules]), 500));
  }
  return apiClient.get('/schedules');
};

// 根據 ID 獲取單一排程
export const getScheduleById = (id: string): Promise<Schedule> => {
  if (import.meta.env.VITE_MOCK_ENABLED === 'true') {
    console.log(`%c MOCKING API: getScheduleById (id: ${id})`, 'color: #00b300');
    const schedule = mockSchedules.find(s => s.id === id);
    return new Promise((resolve, reject) => setTimeout(() => schedule ? resolve(schedule) : reject(new Error('Schedule not found')), 300));
  }
  return apiClient.get(`/schedules/${id}`);
};

// 新增排程
export const createSchedule = (data: Partial<Schedule>): Promise<Schedule> => {
  if (import.meta.env.VITE_MOCK_ENABLED === 'true') {
    console.log('%c MOCKING API: createSchedule', 'color: #00b300', data);
    const newSchedule: Schedule = {
      id: `sched-${Date.now()}`,
      name: data.name || 'New Mocked Schedule',
      cron_spec: data.cron_spec || '0 0 * * *',
      timezone: data.timezone || 'UTC',
      recipients: data.recipients || { to: [] },
      report_ids: data.report_ids || [],
      is_enabled: data.is_enabled || false,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    mockSchedules.push(newSchedule);
    return new Promise(resolve => setTimeout(() => resolve(newSchedule), 500));
  }
  return apiClient.post('/schedules', data);
};

// 更新排程
export const updateSchedule = (id: string, data: Partial<Schedule>): Promise<{ message: string }> => {
  if (import.meta.env.VITE_MOCK_ENABLED === 'true') {
    console.log(`%c MOCKING API: updateSchedule (id: ${id})`, 'color: #00b300', data);
    const index = mockSchedules.findIndex(s => s.id === id);
    if (index !== -1) {
        mockSchedules[index] = { ...mockSchedules[index], ...data, updated_at: new Date().toISOString() };
    }
    return new Promise(resolve => setTimeout(() => resolve({ message: '排程已成功更新' }), 500));
  }
  return apiClient.put(`/schedules/${id}`, data);
};

// 刪除排程
export const deleteSchedule = (id: string): Promise<{ message: string }> => {
  if (import.meta.env.VITE_MOCK_ENABLED === 'true') {
    console.log(`%c MOCKING API: deleteSchedule (id: ${id})`, 'color: #00b300');
    return new Promise(resolve => setTimeout(() => resolve({ message: '排程已成功刪除' }), 500));
  }
  return apiClient.delete(`/schedules/${id}`);
};

// 手動觸發排程
export const triggerSchedule = (id: string): Promise<{ message: string; task_id: string }> => {
  if (import.meta.env.VITE_MOCK_ENABLED === 'true') {
    console.log(`%c MOCKING API: triggerSchedule (id: ${id})`, 'color: #00b300');
    const taskId = `mock-task-${Date.now()}`;
    return new Promise(resolve => setTimeout(() => resolve({ message: '已成功觸發測試任務', task_id: taskId }), 1000));
  }
  return apiClient.post(`/schedules/${id}/trigger`);
};
