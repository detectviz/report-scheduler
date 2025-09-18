import React, { useState, useEffect } from 'react';
import { Form, Input, Select, DatePicker, Divider, Transfer, Button, message, Typography } from 'antd';

const { Title } = Typography;
const { RangePicker } = DatePicker;

const ReportDefinitionPage: React.FC = () => {
    const [targetKeys, setTargetKeys] = useState<string[]>(['1', '5']);
    const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
    const [form] = Form.useForm();

    const [dataSources, setDataSources] = useState<any[]>([]);
    const [loadingDataSources, setLoadingDataSources] = useState(true);

    useEffect(() => {
        setLoadingDataSources(true);
        // FIXME: Replace with actual API call
        const mockDataSources = [
            { id: '1', name: '公司正式環境 Kibana (verified)', type: 'kibana', status: 'verified' },
            { id: '2', name: '測試環境 Grafana (unverified)', type: 'grafana', status: 'unverified' },
        ];
        setDataSources(mockDataSources);
        setLoadingDataSources(false);
    }, []);

    const mockReportElements = Array.from({ length: 20 }).map((_, i) => ({
      key: i.toString(),
      title: `儀表板或圖表 ${i + 1}`,
      description: `這是項目 ${i + 1} 的描述`,
    }));

    const onTransferChange = (nextTargetKeys: string[]) => {
        setTargetKeys(nextTargetKeys);
    };

    const onTransferSelectChange = (sourceSelectedKeys: string[], targetSelectedKeys: string[]) => {
        setSelectedKeys([...sourceSelectedKeys, ...targetSelectedKeys]);
    };

    const [isSaving, setIsSaving] = useState(false);

    const onSave = () => {
        form.validateFields().then(values => {
            setIsSaving(true);
            console.log("Form values:", values);
            console.log("Selected elements:", targetKeys);
            // FIXME: Implement actual API call
            message.success('報表定義已成功儲存！');
            setIsSaving(false);
        }).catch(info => {
            console.log('Validate Failed:', info);
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
                oneWay
                />
            <Divider />
             <Form.Item>
                <Button type="primary" htmlType="submit" loading={isSaving}>儲存報表定義</Button>
            </Form.Item>
        </Form>
    );
};

export default ReportDefinitionPage;
