import React, { useState, useEffect } from 'react';
import { Button, Modal, Form, Input, Switch, message, Table, Space } from 'antd';

const ScheduleManagementPage: React.FC = () => {
    const [isModalVisible, setIsModalVisible] = useState(false);
    const [editingRecord, setEditingRecord] = useState<any | null>(null);
    const [form] = Form.useForm();
    const [schedules, setSchedules] = useState([
      {
        key: '1',
        id: '1',
        name: '每日營運日報',
        cron_spec: '0 9 * * 1-5',
        recipients: 'ops-team@mycompany.com',
        is_enabled: true,
      },
      {
        key: '2',
        id: '2',
        name: '每週產品銷售報表',
        cron_spec: '0 10 * * 1',
        recipients: 'sales@mycompany.com, product@mycompany.com',
        is_enabled: true,
      },
      {
        key: '3',
        id: '3',
        name: '月底財務結算報表',
        cron_spec: '0 22 L * *',
        recipients: 'finance-dept@mycompany.com',
        is_enabled: false,
      },
    ]);

    const handleSwitchChange = (checked: boolean, record: any) => {
        const newSchedules = schedules.map(item => {
            if (item.key === record.key) {
                return { ...item, is_enabled: checked };
            }
            return item;
        });
        setSchedules(newSchedules);
        // FIXME: Add API call to update status
        message.success(`排程 "${record.name}" 已${checked ? '啟用' : '停用'}`);
    };

    useEffect(() => {
        if (editingRecord) {
            form.setFieldsValue(editingRecord);
        } else {
            form.resetFields();
        }
    }, [editingRecord, form]);

    const showModal = (record: any | null = null) => {
        setEditingRecord(record);
        setIsModalVisible(true);
    };

    const handleOk = () => {
        form.validateFields().then(values => {
            setIsModalVisible(false);
            // FIXME: Add API call to create/update
            message.success(editingRecord ? '排程已更新' : '排程已新增');
        }).catch(info => {
            console.log('Validate Failed:', info);
        });
    };

    const handleCancel = () => {
        setIsModalVisible(false);
    };

    const handleTest = (record: any) => {
        message.loading({ content: `正在測試排程 "${record.name}"...`, key: 'test' });
        // FIXME: Add API call to trigger test
        setTimeout(() => {
            message.success({ content: `"${record.name}" 測試郵件已寄送！`, key: 'test', duration: 2 });
        }, 1500);
    };

    const handleDelete = (record: any) => {
        Modal.confirm({
            title: `確定要刪除排程 "${record.name}"嗎?`,
            content: '此操作無法復原。',
            okText: '確定',
            okType: 'danger',
            cancelText: '取消',
            onOk() {
                // FIXME: Add API call to delete
                message.success(`排程 "${record.name}" 已刪除`);
            },
        });
    };

    const scheduleColumns = [
        { title: '排程名稱', dataIndex: 'name', key: 'name' },
        { title: 'Cron 表達式', dataIndex: 'cron_spec', key: 'cron_spec', render: (text: string) => <code>{text}</code> },
        { title: '收件者', dataIndex: 'recipients', key: 'recipients' },
        {
            title: '狀態',
            dataIndex: 'is_enabled',
            key: 'is_enabled',
            render: (isEnabled: boolean, record: any) => (
                <Switch checked={isEnabled} onChange={(checked) => handleSwitchChange(checked, record)} />
            )
        },
        {
            title: '操作',
            key: 'action',
            render: (_: any, record: any) => (
                <Space size="middle">
                    <a onClick={() => showModal(record)}>編輯</a>
                    <a onClick={() => handleTest(record)}>立即測試</a>
                    <a onClick={() => handleDelete(record)} style={{color: 'red'}}>刪除</a>
                </Space>
            ),
        },
    ];

    return (
        <div>
            <Button type="primary" onClick={() => showModal()} style={{ marginBottom: 16 }}>
                新增排程
            </Button>
            <Table columns={scheduleColumns} dataSource={schedules} />
            <Modal
                title={editingRecord ? '編輯排程' : '新增排程'}
                open={isModalVisible}
                onOk={handleOk}
                onCancel={handleCancel}
                destroyOnClose
            >
                <Form form={form} layout="vertical" name="scheduleForm">
                    <Form.Item name="name" label="排程名稱" rules={[{ required: true }]}>
                        <Input />
                    </Form.Item>
                    <Form.Item name="cron_spec" label="Cron 表達式" rules={[{ required: true }]}>
                        <Input />
                    </Form.Item>
                    <Form.Item name="recipients" label="收件者 (以逗號分隔)" rules={[{ required: true }]}>
                        <Input.TextArea rows={3} />
                    </Form.Item>
                </Form>
            </Modal>
        </div>
    );
};

export default ScheduleManagementPage;
