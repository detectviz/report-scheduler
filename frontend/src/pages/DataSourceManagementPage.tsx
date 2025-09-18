import React, { useState, useEffect, useCallback } from 'react';
import { Button, Modal, Form, Input, Select, message, Table, Tag, Space, Typography, Popconfirm } from 'antd';
import {
    getDataSources,
    createDataSource,
    updateDataSource,
    deleteDataSource,
    validateDataSource,
} from '../api/dataSource';
import type { DataSource } from '../api/dataSource';

const { Title } = Typography;
const { Option } = Select;

const DataSourceManagementPage: React.FC = () => {
    const [isModalVisible, setIsModalVisible] = useState(false);
    const [editingRecord, setEditingRecord] = useState<DataSource | null>(null);
    const [data, setData] = useState<DataSource[]>([]);
    const [loading, setLoading] = useState(true);
    const [form] = Form.useForm();

    // 根據所選類型，動態顯示/隱藏表單欄位
    const selectedType = Form.useWatch('type', form);

    const fetchData = useCallback(async () => {
        try {
            setLoading(true);
            const apiData = await getDataSources();
            setData(apiData.map((item) => ({ ...item, key: item.id })));
        } catch (error) {
            // 錯誤訊息已由 apiClient 攔截器統一處理
            console.error("Fetch error:", error);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    // 使用 useEffect 來處理表單值的同步，這比在 showModal 中直接設定更穩健
    useEffect(() => {
        if (isModalVisible && editingRecord) {
            form.setFieldsValue(editingRecord);
        } else {
            form.resetFields();
        }
    }, [editingRecord, isModalVisible, form]);

    const showModal = (record: DataSource | null = null) => {
        setEditingRecord(record);
        setIsModalVisible(true);
    };

    const handleCancel = () => {
        setIsModalVisible(false);
    };

    const handleOk = async () => {
        try {
            const values = await form.validateFields();

            if (editingRecord) {
                // 更新時，只傳送表單中的值
                await updateDataSource(editingRecord.id, values);
                message.success('資料來源已成功更新');
            } else {
                // 新增時，傳送表單中的值並加上預設狀態
                const payload = { ...values, status: 'unverified' };
                await createDataSource(payload);
                message.success('資料來源已成功新增');
            }
            setIsModalVisible(false);
            fetchData(); // 重新整理列表
        } catch (info) {
            console.log('Validate Failed:', info);
        }
    };

    const handleDelete = async (id: string) => {
        try {
            await deleteDataSource(id);
            message.success('資料來源已成功刪除');
            fetchData(); // 重新整理列表
        } catch (error) {
            console.error("Delete error:", error);
        }
    };

    const handleValidate = async (record: DataSource) => {
        message.loading({ content: `正在驗證 "${record.name}"...`, key: record.id });
        try {
            const result = await validateDataSource(record.id);
            message.success({ content: result.message, key: record.id, duration: 2 });
            fetchData(); // 重新整理列表以更新狀態
        } catch (error) {
            // 錯誤訊息已由 apiClient 攔截器統一處理
            console.error("Validation error:", error);
            // 即使 apiClient 顯示了錯誤，我們也需要移除 loading 訊息
            message.error({ content: `"${record.name}" 驗證失敗`, key: record.id, duration: 2 });
        }
    };

    const columns = [
        { title: '名稱', dataIndex: 'name', key: 'name', render: (text: string, record: DataSource) => <Button type="link" onClick={() => showModal(record)} style={{ padding: 0 }}>{text}</Button> },
        {
            title: '類型', dataIndex: 'type', key: 'type', render: (type: string) => {
                const color = type === 'kibana' ? 'geekblue' : 'volcano';
                return <Tag color={color}>{type.toUpperCase()}</Tag>;
            }
        },
        { title: 'URL', dataIndex: 'url', key: 'url' },
        { title: '版本', dataIndex: 'version', key: 'version' },
        {
            title: '狀態', dataIndex: 'status', key: 'status', render: (status: string) => {
                let color: string;
                let text: string;
                switch (status) {
                    case 'verified':
                        color = 'success';
                        text = '已驗證';
                        break;
                    case 'unverified':
                        color = 'warning';
                        text = '未驗證';
                        break;
                    default:
                        color = 'error';
                        text = '連線失敗';
                }
                return <Tag color={color}>{text}</Tag>;
            }
        },
        {
            title: '操作', key: 'action', render: (_: unknown, record: DataSource) => (
                <Space size="middle">
                    <Button type="link" onClick={() => showModal(record)} style={{ padding: 0 }}>編輯</Button>
                    <Popconfirm
                        title={`確定要刪除 "${record.name}" 嗎?`}
                        onConfirm={() => handleDelete(record.id)}
                        okText="確定"
                        cancelText="取消"
                    >
                        <Button type="link" danger style={{ padding: 0 }}>刪除</Button>
                    </Popconfirm>
                    <Button type="link" onClick={() => handleValidate(record)} style={{ padding: 0 }}>驗證連線</Button>
                </Space>
            )
        },
    ];

    return (
        <div>
            <Title level={2}>資料來源管理</Title>
            <Button type="primary" onClick={() => showModal()} style={{ marginBottom: 16 }}>新增資料來源</Button>
            <Table columns={columns} dataSource={data} loading={loading} rowKey="id" />
            <Modal
                title={editingRecord ? '編輯資料來源' : '新增資料來源'}
                open={isModalVisible}
                onOk={handleOk}
                onCancel={handleCancel}
                destroyOnClose
                forceRender
            >
                <Form form={form} layout="vertical" name="dataSourceForm" initialValues={{ type: null }}>
                    <Form.Item name="name" label="名稱" rules={[{ required: true, message: '請輸入名稱' }]}><Input /></Form.Item>
                    <Form.Item name="type" label="類型" rules={[{ required: true, message: '請選擇類型' }]}>
                        <Select placeholder="請選擇資料來源類型" disabled={!!editingRecord}>
                            <Option value="kibana">Kibana</Option>
                            <Option value="grafana">Grafana</Option>
                        </Select>
                    </Form.Item>

                    {selectedType === 'kibana' && (
                        <>
                            <Form.Item name="url" label="Kibana URL" rules={[{ required: true, message: '請輸入 Kibana URL' }]}>
                                <Input placeholder="https://kibana.mycompany.com" />
                            </Form.Item>
                            <Form.Item name="api_url" label="Elasticsearch API URL" rules={[{ required: true, message: '請輸入 Elasticsearch API URL' }]}>
                                <Input placeholder="https://es.mycompany.com:9200" />
                            </Form.Item>
                        </>
                    )}

                    {selectedType === 'grafana' && (
                        <Form.Item name="url" label="Grafana URL" rules={[{ required: true, message: '請輸入 Grafana URL' }]}>
                            <Input placeholder="https://grafana.mycompany.com" />
                        </Form.Item>
                    )}

                    <Form.Item name="auth_type" label="認證方式" rules={[{ required: true, message: '請選擇認證方式' }]}>
                        <Select placeholder="請選擇認證方式">
                            <Option value="basic_auth">基本認證 (帳號密碼)</Option>
                            <Option value="api_token">API Token</Option>
                        </Select>
                    </Form.Item>

                    <Form.Item name="version" label="版本 (選填)"><Input /></Form.Item>
                </Form>
            </Modal>
        </div>
    );
};

export default DataSourceManagementPage;
