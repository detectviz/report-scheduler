import React, { useState } from 'react';
import { Table, Tag, Space, Button, Modal, message, Descriptions } from 'antd';

const HistoryPage: React.FC = () => {
    const [isDetailModalVisible, setIsDetailModalVisible] = useState(false);
    const [selectedRecord, setSelectedRecord] = useState<any | null>(null);

    const handleResend = (record: any) => {
         message.loading({ content: `正在重送紀錄 "${record.scheduleName}"...`, key: 'resend' });
        // FIXME: Add API call to resend
        setTimeout(() => {
            message.success({ content: `紀錄 "${record.scheduleName}" 已成功重送！`, key: 'resend', duration: 2 });
        }, 1500);
    };

    const showDetails = (record: any) => {
        setSelectedRecord(record);
        setIsDetailModalVisible(true);
    };

    const handleDetailModalClose = () => {
        setIsDetailModalVisible(false);
        setSelectedRecord(null);
    };

    const mockHistoryData = [
      {
        key: '1',
        id: '1',
        scheduleName: '每日營運日報',
        triggerTime: '2023-10-27 09:00:01',
        duration: '35s',
        status: 'success',
        recipients: 'ops-team@mycompany.com',
        reportURL: 'https://storage.minio/reports/report-1.pdf'
      },
      {
        key: '2',
        id: '2',
        scheduleName: '每週產品銷售報表',
        triggerTime: '2023-10-27 08:00:05',
        duration: '1m 12s',
        status: 'success',
        recipients: 'sales@mycompany.com',
        reportURL: 'https://storage.minio/reports/report-2.pdf'
      },
      {
        key: '3',
        id: '3',
        scheduleName: '每日營運日報',
        triggerTime: '2023-10-26 09:00:02',
        duration: '2m 5s',
        status: 'error',
        errorMessage: 'Kibana server timeout: Failed to fetch data from dashboard [dashboard-id]',
        recipients: 'ops-team@mycompany.com',
        reportURL: null
      },
    ];

    const historyColumns = [
        { title: '排程名稱', dataIndex: 'scheduleName', key: 'scheduleName' },
        { title: '觸發時間', dataIndex: 'triggerTime', key: 'triggerTime' },
        { title: '執行耗時', dataIndex: 'duration', key: 'duration' },
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
            render: (_: any, record: any) => (
                <Space size="middle">
                    <a onClick={() => showDetails(record)}>查看詳情</a>
                    <a onClick={() => handleResend(record)}>重寄</a>
                </Space>
            ),
        },
    ];

    return (
        <div>
            <Table columns={historyColumns} dataSource={mockHistoryData} />
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
                        <Descriptions.Item label="排程名稱">{selectedRecord.scheduleName}</Descriptions.Item>
                        <Descriptions.Item label="觸發時間">{selectedRecord.triggerTime}</Descriptions.Item>
                        <Descriptions.Item label="執行耗時">{selectedRecord.duration}</Descriptions.Item>
                        <Descriptions.Item label="狀態">
                            <Tag color={selectedRecord.status === 'success' ? 'success' : 'error'}>
                                {selectedRecord.status.toUpperCase()}
                            </Tag>
                        </Descriptions.Item>
                        <Descriptions.Item label="收件者">{selectedRecord.recipients}</Descriptions.Item>
                        {selectedRecord.status === 'error' && (
                             <Descriptions.Item label="錯誤訊息">{selectedRecord.errorMessage}</Descriptions.Item>
                        )}
                        {selectedRecord.reportURL && (
                             <Descriptions.Item label="報表連結"><a>{selectedRecord.reportURL}</a></Descriptions.Item>
                        )}
                    </Descriptions>
                )}
            </Modal>
        </div>
    );
};

export default HistoryPage;
