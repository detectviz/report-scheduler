import type { DataSource } from './dataSource';
import type { ReportDefinition } from './report';
import type { Schedule } from './schedule';
import type { HistoryLog } from './history';

export const mockDataSources: DataSource[] = [
    { id: 'ds-1', name: '正式環境 Kibana', type: 'kibana', url: 'https://kibana.prod.com', api_url: 'https://es.prod.com:9200', auth_type: 'api_token', status: 'verified', created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
    { id: 'ds-2', name: '測試環境 Grafana', type: 'grafana', url: 'https://grafana.dev.com', auth_type: 'basic_auth', status: 'unverified', created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
    { id: 'ds-3', name: '本地 Elasticsearch', type: 'kibana', url: 'http://localhost:5601', api_url: 'http://localhost:9200', auth_type: 'none', status: 'error', created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
    { id: 'ds-4', name: '公開 Kibana 體驗環境', type: 'kibana', url: 'https://demo.elastic.co', api_url: 'https://demo.elastic.co', auth_type: 'none', status: 'verified', created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
];

export const mockReportDefinitions: ReportDefinition[] = [
    {
        id: 'report-1',
        name: 'Elastic Agent 狀態儀表板',
        datasource_id: 'ds-4',
        time_range: 'now-7d',
        elements: [{
            id: 'elastic_agent-0600ffa0-6b5e-11ed-98de-67bdecd21824',
            type: 'dashboard',
            title: 'Elastic Agent aashboard',
            order: 1,
        }],
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
    },
    { id: 'report-2', name: '每週伺服器效能監控', datasource_id: 'ds-2', time_range: 'now-7d', elements: [], created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
    { id: 'report-3', name: '本地測試儀表板', datasource_id: 'ds-3', time_range: 'now-1h', elements: [], created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
];

export const mockSchedules: Schedule[] = [
    { id: 'sched-1', name: '每日營運日報', cron_spec: '0 9 * * 1-5', timezone: 'Asia/Taipei', recipients: { to: ['manager@example.com', 'team@example.com'] }, report_ids: ['report-1'], is_enabled: true, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
    { id: 'sched-2', name: '週末伺服器報告', cron_spec: '0 10 * * 6,0', timezone: 'UTC', recipients: { to: ['devops@example.com'] }, report_ids: ['report-2'], is_enabled: false, created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
];

export const mockHistoryLogs: HistoryLog[] = [
    { id: 'log-1', schedule_id: 'sched-1', schedule_name: '每日營運日報', trigger_time: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(), execution_duration_ms: 15300, status: 'success', recipients: '{"to":["manager@example.com"]}', report_url: '/mock-report-1.pdf' },
    { id: 'log-2', schedule_id: 'sched-1', schedule_name: '每日營運日報', trigger_time: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(), execution_duration_ms: 25000, status: 'error', error_message: 'Kibana timeout', recipients: '{"to":["manager@example.com"]}' },
    { id: 'log-3', schedule_id: 'sched-2', schedule_name: '週末伺服器報告', trigger_time: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(), execution_duration_ms: 5000, status: 'success', recipients: '{"to":["devops@example.com"]}', report_url: '/mock-report-2.pdf' },
];
