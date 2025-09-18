import React, { useState, useEffect, useCallback } from 'react';
import { Table, Tag, Space, Button, Modal, message, Descriptions, Spin, Alert } from 'antd';
import { getHistoryByScheduleId, resendHistory } from '../api/history';
import type { HistoryLog } from '../api/history';

// TODO: 這個頁面應該從路由或 props 接收 scheduleId
const MOCK_SCHEDULE_ID = "c1b0a69a-3f8b-4a6d-8f9a-0b1c2d3e4f5g"; // 假設的 ID，用於開發

const HistoryPage: React.FC = () => {
    const [historyData, setHistoryData] = useState<HistoryLog[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);
    const [isDetailModalVisible, setIsDetailModalVisible] = useState(false);
    const [selectedRecord, setSelectedRecord] = useState<HistoryLog | null>(null);

    const fetchHistory = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            // 在實際應用中，scheduleId 應該來自於 props 或路由參數
            const data = await getHistoryByScheduleId(MOCK_SCHEDULE_ID);
            setHistoryData(data);
        } catch (err: any) {
            // 錯誤已由 apiClient 攔截器處理，這裡可選擇性地在 UI 上顯示更多資訊
            setError('無法載入歷史紀錄。');
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchHistory();
    }, [fetchHistory]);

    const handleResend = async (record: HistoryLog) => {
        const key = 'resend';
        message.loading({ content: `正在重送紀錄 "${record.schedule_name}"...`, key });
        try {
            await resendHistory(record.id);
            message.success({ content: `紀錄 "${record.schedule_name}" 已成功加入重送佇列！`, key, duration: 2 });
            // 延遲一秒後重新整理列表，讓後端有時間處理
            setTimeout(fetchHistory, 1000);
        } catch (err) {
            // 錯誤已由 apiClient 攔截器中的 message.error 處理
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
                    <a onClick={() => showDetails(record)}>查看詳情</a>
                    <a onClick={() => handleResend(record)}>重寄</a>
                </Space>
            ),
        },
    ];

    if (loading) {
        return <Spin tip="正在載入歷史紀錄..." />;
    }

    if (error) {
        return <Alert message="錯誤" description={error} type="error" showIcon />;
    }

    return (
        <div>
            <Table columns={historyColumns} dataSource={historyData} />
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
