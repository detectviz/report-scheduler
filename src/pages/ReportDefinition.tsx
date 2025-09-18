import React, { useState, useEffect } from 'react';
import { Form, Input, Select, DatePicker, Divider, Transfer, Button, message, Typography } from 'antd';

const { Title } = Typography;
const { RangePicker } = DatePicker;

// --- TypeScript Interfaces ---
interface DataSource {
    id: string;
    key: string; // for Ant Design Table
    name: string;
    type: 'kibana' | 'grafana';
    status: 'verified' | 'unverified' | 'error';
}

interface ReportElement {
    key: string;
    title: string;
    description: string;
}

interface FormValues {
    reportName: string;
    dataSourceId: string;
    timeRange: [any, any];
}

const ReportDefinitionPage: React.FC = () => {
    const [form] = Form.useForm<FormValues>();
    const [targetKeys, setTargetKeys] = useState<React.Key[]>([]);
    const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);
    const [isSaving, setIsSaving] = useState(false);

    const [dataSources, setDataSources] = useState<DataSource[]>([]);
    const [loadingDataSources, setLoadingDataSources] = useState(true);

    const [mockReportElements, setMockReportElements] = useState<ReportElement[]>([]);

    useEffect(() => {
        setLoadingDataSources(true);
        fetch("/api/v1/datasources")
            .then(res => res.json())
            .then((apiData: DataSource[]) => {
                const processedData = (apiData || []).map(item => ({ ...item, key: item.id }));
                setDataSources(processedData);
            })
            .catch(() => message.error('無法獲取資料來源列表'))
            .finally(() => setLoadingDataSources(false));
    }, []);

    useEffect(() => {
        const elements = Array.from({ length: 20 }).map((_, i) => ({
            key: i.toString(),
            title: `儀表板或圖表 ${i + 1}`,
            description: `這是項目 ${i + 1} 的描述`,
        }));
        setMockReportElements(elements);
    }, []);

    const onTransferChange = (nextTargetKeys: React.Key[]) => {
        setTargetKeys(nextTargetKeys);
    };

    const onTransferSelectChange = (sourceSelectedKeys: React.Key[], targetSelectedKeys: React.Key[]) => {
        setSelectedKeys([...sourceSelectedKeys, ...targetSelectedKeys]);
    };

    const onSave = (values: FormValues) => {
        setIsSaving(true);
        const payload = {
            name: values.reportName,
            description: "由前端應用程式建立的報表",
            datasource_id: values.dataSourceId,
            time_range: "now-1h", // Simplified for now
            elements: targetKeys.map((key) => {
                const element = mockReportElements.find(e => e.key === key);
                return {
                    id: element?.key || '',
                    type: "dashboard",
                    title: element?.title || '無標題元素'
                };
            })
        };

        fetch('/api/v1/reports', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        })
        .then(res => {
            if (res.ok) {
                message.success('報表定義已成功儲存！');
                form.resetFields();
                setTargetKeys([]);
            } else {
                message.error('後端儲存失敗，請檢查日誌');
            }
        })
        .catch(err => {
            message.error(`網路請求失敗: ${err.message}`);
        })
        .finally(() => {
            setIsSaving(false);
        });
    };

    return (
        <Form form={form} layout="vertical" onFinish={onSave}>
            <Title level={4}>基本資訊</Title>
            <Form.Item label="報表名稱" name="reportName" rules={[{ required: true, message: '請輸入報表名稱' }]}>
                <Input placeholder="例如：每日網站流量分析報表" />
            </Form.Item>
            <Form.Item label="選擇資料來源" name="dataSourceId" rules={[{ required: true, message: '請選擇一個資料來源' }]}>
                <Select placeholder="選擇一個已驗證的資料來源" loading={loadingDataSources}>
                    {dataSources.map(ds => (
                        <Select.Option key={ds.id} value={ds.id} disabled={ds.status !== 'verified'}>
                            {ds.name} ({ds.type.toUpperCase()})
                        </Select.Option>
                    ))}
                </Select>
            </Form.Item>
            <Form.Item label="設定時間範圍" name="timeRange" rules={[{ required: true, message: '請設定時間範圍' }]}>
                <RangePicker style={{ width: '100%' }} showTime />
            </Form.Item>
            <Divider />
            <Title level={4}>挑選報表元素</Title>
            <p>請從左側選擇您想包含在此報表中的儀表板或圖表，並可拖曳右側項目進行排序。</p>
            <Transfer
                dataSource={mockReportElements}
                titles={['可選項目', '已選項目']}
                targetKeys={targetKeys}
                selectedKeys={selectedKeys}
                onChange={onTransferChange}
                onSelectChange={onTransferSelectChange}
                render={item => item.title}
                listStyle={{
                    width: '100%',
                    height: 300,
                }}
            />
            <Divider />
            <Form.Item>
                <Button type="primary" htmlType="submit" loading={isSaving}>儲存報表定義</Button>
            </Form.Item>
        </Form>
    );
};

export default ReportDefinitionPage;
