import React, { useState, useEffect, useCallback } from 'react';
import { Button, Table, Space, Typography, Popconfirm, message } from 'antd';
import type { TableProps } from 'antd';
import { useNavigate } from 'react-router-dom';
import { getReportDefinitions, deleteReportDefinition, generateReportPreview } from '../api/report';
import type { ReportDefinition, ReportElement } from '../api/report';
import { getDataSources } from '../api/dataSource';
import type { DataSource } from '../api/dataSource';

const { Title } = Typography;

const ReportDefinitionPage: React.FC = () => {
    const navigate = useNavigate();
    const [reports, setReports] = useState<ReportDefinition[]>([]);
    const [dataSources, setDataSources] = useState<Record<string, DataSource>>({});
    const [loading, setLoading] = useState(true);
    const [previewLoading, setPreviewLoading] = useState<Record<string, boolean>>({});

    const fetchData = useCallback(async () => {
        setLoading(true);
        try {
            const [reportsData, dataSourcesData] = await Promise.all([
                getReportDefinitions(),
                getDataSources(),
            ]);

            const dataSourceMap = dataSourcesData.reduce((acc, ds) => {
                acc[ds.id] = ds;
                return acc;
            }, {} as Record<string, DataSource>);

            setReports(reportsData);
            setDataSources(dataSourceMap);

        } catch (error) {
            console.error("Failed to fetch data:", error);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    const handleAddReport = () => {
        navigate('/reports/new');
    };

    const handleEditReport = (id: string) => {
        navigate(`/reports/edit/${id}`);
    };

    const handleDeleteReport = async (id: string) => {
        try {
            await deleteReportDefinition(id);
            message.success('報表定義已成功刪除');
            fetchData(); // Refresh data
        } catch (error) {
            console.error("Failed to delete report definition:", error);
        }
    };

    const handlePreview = async (record: ReportDefinition) => {
        setPreviewLoading(prev => ({ ...prev, [record.id]: true }));
        try {
            const result = await generateReportPreview(record.id);
            if (result.preview_url) {
                // 在本地開發時，URL 可能是相對的，需要加上 base URL
                const baseUrl = import.meta.env.VITE_MOCK_ENABLED === 'true' ? '' : window.location.origin;
                window.open(baseUrl + result.preview_url, '_blank');
            }
        } catch (error) {
            // 錯誤已由 apiClient 處理
        } finally {
            setPreviewLoading(prev => ({ ...prev, [record.id]: false }));
        }
    };

    const columns: TableProps<ReportDefinition>['columns'] = [
        {
            title: '報表名稱',
            dataIndex: 'name',
            key: 'name',
            render: (text: string, record: ReportDefinition) => <Button type="link" onClick={() => handleEditReport(record.id)} style={{ padding: 0 }}>{text}</Button>,
        },
        {
            title: '資料來源',
            dataIndex: 'datasource_id',
            key: 'datasource_id',
            render: (dataSourceId: string) => dataSources[dataSourceId]?.name || '未知',
        },
        {
            title: '元素數量',
            dataIndex: 'elements',
            key: 'elements',
            align: 'center',
            render: (elements: ReportElement[]) => (elements || []).length,
        },
        {
            title: '操作',
            key: 'action',
            render: (_: unknown, record: ReportDefinition) => (
                <Space size="middle">
                    <Button type="link" onClick={() => handleEditReport(record.id)} style={{ padding: 0 }}>編輯</Button>
                    <Popconfirm
                        title={`確定要刪除 "${record.name}" 嗎?`}
                        onConfirm={() => handleDeleteReport(record.id)}
                        okText="確定"
                        cancelText="取消"
                    >
                        <Button type="link" danger style={{ padding: 0 }}>刪除</Button>
                    </Popconfirm>
                    <Button
                        type="link"
                        style={{ padding: 0 }}
                        onClick={() => handlePreview(record)}
                        loading={previewLoading[record.id]}
                    >
                        立即執行與預覽
                    </Button>
                </Space>
            ),
        },
    ];

    return (
        <div>
            <Title level={2}>報表定義</Title>
            <p>管理所有已設定的報表定義。您可以在此建立新報表、編輯現有設定或手動觸發執行。</p>
            <Button type="primary" onClick={handleAddReport} style={{ marginBottom: 16 }}>
                新增報表定義
            </Button>
            <Table
                columns={columns}
                dataSource={reports.map(r => ({ ...r, key: r.id }))}
                loading={loading}
                rowKey="id"
            />
        </div>
    );
};

export default ReportDefinitionPage;
