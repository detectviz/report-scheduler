import React, { useState, useEffect, useCallback } from 'react';
import { Table, Tag, Space, Button, Modal, message, Descriptions, Spin, Alert, Typography } from 'antd';
import { useParams, Link } from 'react-router-dom';
import { getHistoryByScheduleId, resendHistory } from '../api/history';
import type { HistoryLog } from '../api/history';

const { Title } = Typography;

const HistoryPage: React.FC = () => {
    const { scheduleId } = useParams<{ scheduleId: string }>();
    const [historyData, setHistoryData] = useState<HistoryLog[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);
    const [isDetailModalVisible, setIsDetailModalVisible] = useState(false);
    const [selectedRecord, setSelectedRecord] = useState<HistoryLog | null>(null);

    const fetchHistory = useCallback(async () => {
        if (!scheduleId) return;

        setLoading(true);
        setError(null);
        try {
            const data = await getHistoryByScheduleId(scheduleId);
            setHistoryData(data || []);
        } catch (err: any) {
            setError('無法載入歷史紀錄。');
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, [scheduleId]);

    useEffect(() => {
        fetchHistory();
    }, [fetchHistory]);

    const handleResend = async (record: HistoryLog) => {
        const key = 'resend';
        message.loading({ content: `正在重送紀錄 "${record.schedule_name}"...`, key });
        try {
            await resendHistory(record.id);
            message.success({ content: `紀錄 "${record.schedule_name}" 已成功加入重送佇列！`, key, duration: 2 });
            setTimeout(fetchHistory, 1000);
        } catch (err) {
            console.error(err);
        }
    };

    const showDetails = (record: HistoryLog) => {
        setSelectedRecord(record);
        setIsDetailModalVisible(true);
    };

    const handleDetailModalClose = () => {
        setIsDetailModalVisible(false);
        setSelectedRecord(null);
    };

    const historyColumns = [
        { title: '排程名稱', dataIndex: 'schedule_name', key: 'schedule_name' },
        {
            title: '觸發時間',
            dataIndex: 'trigger_time',
            key: 'trigger_time',
            render: (text: string) => new Date(text).toLocaleString(),
        },
        {
            title: '執行耗時',
            dataIndex: 'execution_duration_ms',
            key: 'execution_duration_ms',
            render: (ms: number) => `${(ms / 1000).toFixed(2)}s`
        },
        {
            title: '狀態',
            dataIndex: 'status',
            key: 'status',
            render: (status: string) => {
                let color = status === 'success' ? 'success' : 'error';
                return <Tag color={color}>{status.toUpperCase()}</Tag>;
            }
        },
        {
            title: '操作',
            key: 'action',
            render: (_: any, record: HistoryLog) => (
                <Space size="middle">
                    <Button type="link" style={{ padding: 0 }} onClick={() => showDetails(record)}>查看詳情</Button>
                    <Button type="link" style={{ padding: 0 }} onClick={() => handleResend(record)}>重寄</Button>
                </Space>
            ),
        },
    ];

    if (!scheduleId) {
        return (
            <Alert
                message="錯誤"
                description={<span>無效的排程 ID。請從 <Link to="/schedules">排程管理</Link> 頁面進入。</span>}
                type="error"
                showIcon
            />
        );
    }

    if (loading) {
        return <Spin tip="正在載入歷史紀錄..." />;
    }

    if (error) {
        return <Alert message="錯誤" description={error} type="error" showIcon />;
    }

    return (
        <div>
            <Title level={2}>執行歷史紀錄</Title>
            <Table columns={historyColumns} dataSource={historyData} rowKey="id" />
            <Modal
                title="執行紀錄詳情"
                open={isDetailModalVisible}
                onCancel={handleDetailModalClose}
                footer={[
                    <Button key="back" onClick={handleDetailModalClose}>
                      關閉
                    </Button>,
                  ]}
            >
                {selectedRecord && (
                    <Descriptions bordered column={1}>
                        <Descriptions.Item label="排程名稱">{selectedRecord.schedule_name}</Descriptions.Item>
                        <Descriptions.Item label="觸發時間">{new Date(selectedRecord.trigger_time).toLocaleString()}</Descriptions.Item>
                        <Descriptions.Item label="執行耗時">{`${(selectedRecord.execution_duration_ms / 1000).toFixed(2)}s`}</Descriptions.Item>
                        <Descriptions.Item label="狀態">
                            <Tag color={selectedRecord.status === 'success' ? 'success' : 'error'}>
                                {selectedRecord.status.toUpperCase()}
                            </Tag>
                        </Descriptions.Item>
                        <Descriptions.Item label="收件者">{selectedRecord.recipients}</Descriptions.Item>
                        {selectedRecord.status === 'error' && (
                             <Descriptions.Item label="錯誤訊息">{selectedRecord.error_message}</Descriptions.Item>
                        )}
                        {selectedRecord.report_url && (
                             <Descriptions.Item label="報表連結"><a>{selectedRecord.report_url}</a></Descriptions.Item>
                        )}
                    </Descriptions>
                )}
            </Modal>
        </div>
    );
};

export default HistoryPage;
