import React, { useState, useEffect, useCallback } from 'react';
import { Button, Modal, Form, Input, Switch, message, Table, Space, Select, Typography, Popconfirm } from 'antd';
import { getSchedules, createSchedule, updateSchedule, deleteSchedule, triggerSchedule } from '../api/schedule';
import type { Schedule } from '../api/schedule';
import { getReportDefinitions } from '../api/report';
import type { ReportDefinition } from '../api/report';

const { Title } = Typography;
const { Option } = Select;

const timezones = ["UTC", "Asia/Taipei", "Asia/Tokyo", "America/New_York", "Europe/London"];

const ScheduleManagementPage: React.FC = () => {
    const [isModalVisible, setIsModalVisible] = useState(false);
    const [editingRecord, setEditingRecord] = useState<Schedule | null>(null);
    const [form] = Form.useForm();
    const [schedules, setSchedules] = useState<Schedule[]>([]);
    const [reportDefinitions, setReportDefinitions] = useState<ReportDefinition[]>([]);
    const [loading, setLoading] = useState(true);

    const fetchData = useCallback(async () => {
        setLoading(true);
        try {
            const [schedulesData, reportsData] = await Promise.all([getSchedules(), getReportDefinitions()]);
            setSchedules(schedulesData);
            setReportDefinitions(reportsData);
        } catch {
            // Error is handled by the apiClient interceptor
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    const handleSwitchChange = async (checked: boolean, record: Schedule) => {
        try {
            await updateSchedule(record.id, { ...record, is_enabled: checked });
            message.success(`排程 "${record.name}" 已${checked ? '啟用' : '停用'}`);
            fetchData();
        } catch {
            message.error('狀態更新失敗');
        }
    };

    useEffect(() => {
        if (isModalVisible && editingRecord) {
            form.setFieldsValue({
                ...editingRecord,
                recipients_to: editingRecord.recipients?.to || [],
            });
        } else {
            form.resetFields();
        }
    }, [editingRecord, isModalVisible, form]);

    const showModal = (record: Schedule | null = null) => {
        setEditingRecord(record);
        setIsModalVisible(true);
    };

    const handleOk = async () => {
        try {
            const values = await form.validateFields();
            const payload = {
                ...values,
                recipients: { to: values.recipients_to || [] },
            };
            delete payload.recipients_to;

            if (editingRecord) {
                await updateSchedule(editingRecord.id, payload);
                message.success('排程已成功更新');
            } else {
                await createSchedule(payload);
                message.success('排程已成功新增');
            }
            setIsModalVisible(false);
            fetchData();
        } catch (info) {
            console.log('Validate Failed:', info);
        }
    };

    const handleCancel = () => {
        setIsModalVisible(false);
    };

    const handleTest = async (record: Schedule) => {
        message.loading({ content: `正在觸發排程 "${record.name}"...`, key: record.id });
        try {
            const result = await triggerSchedule(record.id);
            message.success({ content: `${result.message} (Task ID: ${result.task_id})`, key: record.id, duration: 3 });
        } catch {
            message.error({ content: `觸發失敗`, key: record.id, duration: 2 });
        }
    };

    const handleDelete = async (id: string) => {
        try {
            await deleteSchedule(id);
            message.success(`排程已刪除`);
            fetchData();
        } catch {
            // Error is handled by the apiClient interceptor
        }
    };

    const scheduleColumns = [
        { title: '排程名稱', dataIndex: 'name', key: 'name' },
        { title: 'Cron 表達式', dataIndex: 'cron_spec', key: 'cron_spec', render: (text: string) => <code>{text}</code> },
        { title: '收件者 (To)', dataIndex: 'recipients', key: 'recipients', render: (recipients: { to?: string[] }) => recipients?.to?.join(', ') || '' },
        {
            title: '狀態',
            dataIndex: 'is_enabled',
            key: 'is_enabled',
            render: (isEnabled: boolean, record: Schedule) => (
                <Switch checked={isEnabled} onChange={(checked) => handleSwitchChange(checked, record)} />
            )
        },
        {
            title: '操作',
            key: 'action',
            render: (_: unknown, record: Schedule) => (
                <Space size="middle">
                    <a onClick={() => showModal(record)}>編輯</a>
                    <a onClick={() => handleTest(record)}>立即測試</a>
                    <Popconfirm
                        title={`確定要刪除排程 "${record.name}" 嗎?`}
                        onConfirm={() => handleDelete(record.id)}
                        okText="確定"
                        cancelText="取消"
                    >
                        <a style={{ color: 'red' }}>刪除</a>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <div>
            <Title level={2}>排程管理</Title>
            <p>集中管理所有報表的寄送排程。您可以在此建立、編輯、啟用/停用或立即測試一個排程。</p>
            <Button type="primary" onClick={() => showModal()} style={{ marginBottom: 16 }}>
                新增排程
            </Button>
            <Table columns={scheduleColumns} dataSource={schedules.map(s => ({ ...s, key: s.id }))} loading={loading} />
            <Modal
                title={editingRecord ? '編輯排程' : '新增排程'}
                open={isModalVisible}
                onOk={handleOk}
                onCancel={handleCancel}
                destroyOnHidden
                width={600}
            >
                <Form form={form} layout="vertical" name="scheduleForm" initialValues={{ timezone: 'Asia/Taipei', is_enabled: true }}>
                    <Form.Item name="name" label="排程名稱" rules={[{ required: true, message: '請輸入排程名稱' }]}>
                        <Input />
                    </Form.Item>
                    <Form.Item name="cron_spec" label="Cron 表達式" rules={[{ required: true, message: '請輸入有效的 Cron 表達式' }]}>
                        <Input placeholder="例如：0 9 * * 1-5 (週一至週五早上 9 點)" />
                    </Form.Item>
                    <Form.Item name="timezone" label="時區" rules={[{ required: true }]}>
                        <Select>
                            {timezones.map(tz => <Option key={tz} value={tz}>{tz}</Option>)}
                        </Select>
                    </Form.Item>
                    <Form.Item name="report_ids" label="選擇報表" rules={[{ required: true, message: '請至少選擇一份報表' }]}>
                        <Select mode="multiple" placeholder="選擇要附加在此排程的報表" loading={loading}>
                            {reportDefinitions.map(report => (
                                <Option key={report.id} value={report.id}>{report.name}</Option>
                            ))}
                        </Select>
                    </Form.Item>
                    <Form.Item name="recipients_to" label="收件者 (To)" rules={[{ required: true, message: '請至少輸入一位收件者' }]}>
                        <Select mode="tags" tokenSeparators={[',', ' ']} placeholder="輸入郵件地址後按 Enter" />
                    </Form.Item>
                    <Form.Item name="email_subject" label="郵件主旨">
                        <Input placeholder="[每日報表] {{report_name}} - {{date}}" />
                    </Form.Item>
                    <Form.Item name="email_body" label="郵件內文">
                        <Input.TextArea rows={4} placeholder="您好，附件為今日的營運報表。" />
                    </Form.Item>
                    <Form.Item name="is_enabled" label="啟用狀態" valuePropName="checked">
                        <Switch />
                    </Form.Item>
                </Form>
            </Modal>
        </div>
    );
};

export default ScheduleManagementPage;
