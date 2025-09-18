import React, { useState, useEffect } from 'react';
import { Button, Modal, Form, Input, Select, message, Table, Tag, Space } from 'antd';

const DataSourceManagementPage: React.FC = () => {
    const [isModalVisible, setIsModalVisible] = useState(false);
    const [editingRecord, setEditingRecord] = useState<any | null>(null);
    const [selectedType, setSelectedType] = useState<string | null>(null);
    const [form] = Form.useForm();
    const [data, setData] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);

    const fetchData = () => {
        setLoading(true);
        // FIXME: Replace with actual API call
        // fetch("/api/v1/datasources")
        //     .then(res => res.json())
        //     .then(apiData => {
        //         const tableData = (apiData || []).map((item: any) => ({ ...item, key: item.id }));
        //         setData(tableData);
        //         setLoading(false);
        //     })
        //     .catch(error => {
        //         console.error("Fetch error:", error);
        //         message.error('無法獲取資料來源列表');
        //         setLoading(false);
        //     });

        // Mock Data for now
        const mockData = [
            { id: '1', key: '1', name: '公司正式環境 Kibana', type: 'kibana', url: 'https://kibana.mycompany.com', api_url: 'https://es.mycompany.com:9200', version: '8.5.1', status: 'verified' },
            { id: '2', key: '2', name: '測試環境 Grafana', type: 'grafana', url: 'https://grafana.test.com', version: '9.2.3', status: 'unverified' },
            { id: '3', key: '3', name: '舊版 ES 叢集', type: 'elasticsearch', url: 'http://es-old:9200', version: '7.10.0', status: 'error' },
        ];
        setData(mockData.map(item => ({ ...item, key: item.id })));
        setLoading(false);
    };

    useEffect(() => {
        fetchData();
    }, []);

    useEffect(() => {
        if (editingRecord) {
            form.setFieldsValue(editingRecord);
            setSelectedType(editingRecord.type);
        } else {
            form.resetFields();
            setSelectedType(null);
        }
    }, [editingRecord, form]);

    const showModal = (record: any | null = null) => {
        setEditingRecord(record);
        if (record) {
            setSelectedType(record.type);
        } else {
            setSelectedType(null);
        }
        setIsModalVisible(true);
    };

    const handleOk = () => {
        form.validateFields()
            .then(values => {
                const payload = { ...editingRecord, ...values };
                console.log("Payload:", payload);
                // FIXME: Implement actual API call
                message.success(editingRecord ? '資料來源已更新' : '資料來源已新增');
                setIsModalVisible(false);
                fetchData(); // Refresh the table
            })
            .catch(info => {
                console.log('Validate Failed:', info);
            });
    };

    const handleCancel = () => {
        setIsModalVisible(false);
    };

    const handleValidate = (record: any) => {
        message.loading({ content: `正在驗證 "${record.name}"...`, key: 'validate' });
        setTimeout(() => {
            if (Math.random() > 0.3) {
                message.success({ content: `"${record.name}" 連線成功！`, key: 'validate', duration: 2 });
            } else {
                message.error({ content: `"${record.name}" 連線失敗，請檢查設定。`, key: 'validate', duration: 2 });
            }
        }, 1500);
    };

    const columns = [
        { title: '名稱', dataIndex: 'name', key: 'name', render: (text: string) => <a>{text}</a> },
        { title: '類型', dataIndex: 'type', key: 'type', render: (type: string) => {
            let color = 'geekblue';
            if (type === 'grafana') color = 'volcano';
            if (type === 'elasticsearch') color = 'green';
            return <Tag color={color}>{type.toUpperCase()}</Tag>;
        }},
        { title: 'URL', dataIndex: 'url', key: 'url' },
        { title: '版本', dataIndex: 'version', key: 'version' },
        { title: '狀態', dataIndex: 'status', key: 'status', render: (status: string) => {
            let color: string;
            if (status === 'verified') color = 'success';
            else if (status === 'unverified') color = 'warning';
            else color = 'error';
            const text = status === 'verified' ? '已驗證' : status === 'unverified' ? '未驗證' : '連線失敗';
            return <Tag color={color}>{text}</Tag>;
        }},
        { title: '操作', key: 'action', render: (_: any, record: any) => (
            <Space size="middle">
                <a onClick={() => showModal(record)}>編輯</a>
                <a>刪除</a>
                <a onClick={() => handleValidate(record)}>驗證連線</a>
            </Space>
        )},
    ];

    return (
        <div>
            <Button type="primary" onClick={() => showModal()} style={{ marginBottom: 16 }}>新增資料來源</Button>
            <Table columns={columns} dataSource={data} loading={loading} />
            <Modal title={editingRecord ? '編輯資料來源' : '新增資料來源'} open={isModalVisible} onOk={handleOk} onCancel={handleCancel} destroyOnClose>
                <Form form={form} layout="vertical" name="dataSourceForm" initialValues={{ type: null }}>
                    <Form.Item name="name" label="名稱" rules={[{ required: true, message: '請輸入名稱' }]}><Input /></Form.Item>
                    <Form.Item name="type" label="類型" rules={[{ required: true, message: '請選擇類型' }]}>
                        <Select placeholder="請選擇資料來源類型" onChange={setSelectedType}>
                            <Select.Option value="kibana">Kibana</Select.Option>
                            <Select.Option value="grafana">Grafana</Select.Option>
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

                    <Form.Item name="version" label="版本"><Input /></Form.Item>
                </Form>
            </Modal>
        </div>
    );
};

export default DataSourceManagementPage;
